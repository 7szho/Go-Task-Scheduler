handler 包是项目中执行节点的核心业务逻辑层，负责任务的定义、分发、执行到结果反馈的整个生命周期管理。将模型定义与底层服务粘合起来，实现分布式任务调度的关键所在。
1. 任务执行策略：定义了一个通用的 Handler 接口，并为不同类型的任务（如命令行CMD、HTTP请求）提供了具体的实现
2. 任务处理工厂：提供了一个 CreateHandler 工厂函数，能根据任务类型动态创建对应的处理器实例
3. 任务生命周期管理：实现了任务的完整执行流程，包括：
    - 安全执行与恢复：使用 defer/recover 机制防止单个任务panic影响整个工作节点的稳定性
    - 超时控制：为任务执行提供超时中断机制
    - 重试机制：在任务执行失败后，根据配置多次重试
    - 日志记录：将任务的每一次执行（包括成功、失败、重试）的结果和输出持久化到数据库
    - 失败通知：在任务失败后，通过邮件或WebHook等方式通知相关用户
4. 分布式状态同步：深度整合etcd，用于：
    - 监听任务变化：实时监控etcd中任务的增、删、改事件
    - 管理运行时进程：将正在执行的任务（JobProc）注册到 etcd 并设置租约TTL，作为分布式环境下的心跳和进程管理机制
    - 响应系统指令：监听一次性任务和针对节点的系统级控制指令
---

## 1.接口与工厂  
这部分定义了整个任务执行模块的顶层抽象和对象创建逻辑

#### `Handler`接口
- 作用：定义了所有任务执行器的统一的contract。任何结构体只要实现了 Run 方法，就可以被视为一个种任务处理器。

#### `CreateHandle`函数
- 作用：工厂函数，根据传入的任务类型，生产出对应的具体处理器实例。
- 输入：`j *Job`：一个执行 Job 实例的指针，函数函数根据其 Type 字段来决定创建那种处理器
- 流程：
    1. 初始化一个Handler类型的变量handler为nil
    2. 使用switch语句检查j.Type的值
    3. 如果类型是 models.JobTypeCmd，则创建一个 CMDHandler 的新实例并赋值给 handler
    4. 如果类型是 models.JobTypeHttp，则创建一个 HTTPHandler 的新实例并赋值给 handler
    5. 如果任务类型不匹配任何分支，handler 保持 nil
- 输出：`Hanlder`：一个实现了 Handler 接口的具体处理器实例，或者不匹配时返回 nil

## 2. 任务执行器实现  
这部分是 Handler 接口的具体实现，负责执行不同类型的任务

#### `CMDHandler.Run` 方法
- 作用：实现了 Handler 接口，负责具体执行一个本地命令行任务，并管理其在 etcd 中的生命周期。
- 输入：`job *Job`：需要执行的任务详情
- 流程：
    1. 创建命令：
        - 若 job.Timeout > 0，则创建带超时的 context.Context 并使用 exec.CommandContext 构建命令，以实现超时终止
        - 否则，使用常规的 exec.Command 创建命令
    2. 捕获输出：创建 bytes.Buffer 并同时绑定到 cmd.Stdout 和 cmd.Stderr
    3. 启动命令：调用 cmd.Start() 异步启动
    4. 注册进程：创建 JobProc 实例，并调用 proc.Start() 将其注册到 etcd。使用 defer proc.Stop() 确保任务结束时清理 etcd 记录
    5. 等待结束： 调用 cmd.Wait() 阻塞等待命令执行完成
    6. 处理结果：检查 cmd.Wait() 的返回结果，并记录日志
- 输出：
    1. `result string`：命令的标准输出和标准错误
    2. `err error`：命令启动或执行失败时返回的错误

#### `HTTPHandler.Run` 方法
- 作用：实现了 Handler 接口，负责具体执行一个 HTTP 请求类型的任务
- 输入：`job *Job`：需要执行的任务详情
- 流程：
    1. 注册进程：创建并注册一个 JobProc 到 etcd，将进程ID设为0
    2. 捕获输出：创建 bytes.Buffer 并同时绑定到 cmd.Stdout 和 cmd.Stderr
    3. 启动命令：调用 cmd.Start() 异步启动
- 输出：
    1. `result string`：HTTP 请求的响应体
    2. `err error`：请求过程中的错误

## 3. 任务生命周期管理  
这部分涵盖了从任务获取、执行、重试到日志记录的完整流程。

#### `GetJobs / GetJobAndRev` 函数
- 作用：从 etcd 中获取任务定义。GetJobs 获取指定节点上的全部任务，GetJobAndRev 获取单个任务及其元数据 ModeRevision（可用于乐观锁）
- 输入：
    1. `nodeUUID string`：节点ID
    2. `jobId int`：任务ID
- 流程：
    1. 分别使用 etcdclient.KeyEtcdJobProfile（前缀）和 etcdclient.KeyEtcdJob（精确）构造查询 key
    2. 调用 etcdclient.Get 方法发起查询
    3. 遍历结果，将 JSON 格式的 value 反序列化为 Job 结构体
    4. 调用 job.Check() 和 job.SplitCmd() 验证并预处理数据
- 输出：
    - `GetJobs`：Jobs（即 map[int]*Job）和 error
    - `GetJobAndRev`： *Job、rev int64（即 ModeRevision）和 error

#### `Job.RunWithRecovery` 方法
- 作用：安全地执行单次任务，内置 panic 恢复、日志记录和失败通知机制，通常用于“立即执行”的场景
- 流程：
    1. Panic恢复：使用 defer 和 recover() 捕获执行过程中的 panic，记录堆栈信息，防止主程序崩溃
    2. 创建日志：调用 j.CreateJobLog() 在数据库中创建一个初始日志记录，并捕获 jobLogId
    3. 创建处理器：调用 CreateHandle(j) 获取与任务类型匹配的执行器 h
    4. 执行任务：调用 h.Run(j) 执行任务，并接收返回的 result 和 runErr
    5. 处理结果：
        - 若 runErr 不为 nil（失败）：调用 j.Fail() 更新日志为失败，并异步发送失败通知
        - 若 runErr 为 nil（成功）：调用 j.Success() 更新日志为成功状态
    
#### `CreateJob` 函数
- 作用：将一个 Job 包装成 cron.FuncJob 闭包，以便集成到 cron 调度器中，并内置了完整的重试和通知逻辑。
- 输入：`j *Job`：需要被调度的任务
- 流程：
    1. 创建处理器：在闭包外部预先创建好的任务处理器 h
    2. 返回闭包：返回一个 func()，该函数是 cron 调度器实际执行的内容
    3. 执行与重试：在闭包内内部，使用 for 循环执行任务，总次数为 1 + j.RetryTimes
    4. 成功即退出：如果 h.Run(j) 执行成功，则调用 j.Success() 更新日志并立即退出循环
    5. 失败则等待：如果执行成功，记录警告日志，并根据 j.RetryInterval 或默认递增策略（time.Sleep）进行等待，然后进行下一次重试
    6. 最终失败处理：如果循环结束后任务仍为成功，调用 j.Fail() 将日志最终标记为失败，并构建和发送最终的失败通知。
- 输出：`cron.FuncJob`：一个可直接被 cron 库调度的函数

#### `CreateJobLog, Success, Fail` 任务日志辅助函数
- 作用：封装了对数据库中 job_log 表的 INSERT 和 UPDATE 操作
- CreateJobLog：为一次任务执行在数据库中创建一个初始日志条目，记录任务名、命令、开始时间等信息，并返回新日志的ID。
- Success：一个辅助方法，调用 UpdateJobLog 并将 success 标志位置为 true。
- Fail：一个辅助方法，调用 UpdateJobLog 并将 success 标志位置为 false。

## 4. 运行时进程管理（JobProc）
这部分逻辑用于在分布式环境中追踪正在运行的任务，通过在etcd中创建临时节点实现。

#### `JobProc.Start` 方法
- 作用：在 etcd 中注册一个正在运行的任务进程，并设置租约（TTL）作为心跳机制
- 流程：
    1. 原子性检查：使用 atomic.CompareAndSwapInt32 检查 p.Running 状态，确保注册逻辑只被执行一次
    2. 序列化：使用 json.Marshal 将 p.JobProcVal（包含启动时间等动态信息）序列化为 JSON 字符串
    3. 写入etcd：调用 etcdclient.PutWithTtl，将进程信息写入其唯一的 Key()，并附带一个由系统配置 JobProcTtl 决定的租约

#### `JobProc.Stop` 方法
- 作用：停止对任务进程的追踪，并从 etcd 中清理对应的记录
- 流程：
    1. 原子性检查：使用 atomic.CompareAndSwapInt32 将 p.Running 状态从 1 置为 0，确保清理逻辑只执行一次
    2. 等待同步：调用 p.Wg.Wait() 等待 Start 方法中可能正在进行的 etcd put 操作完成，防止竞争条件
    3. 删除Key：调用内部的 p.del() 方法，即 etcdclient.Delete，来删除 etcd 中的进程记录

## 5. 分布式状态监控（etcd Watchers）
这组函数利用 etcd 的 Watch 机制，实现对任务、进程和系统指令变化的实时监控

#### `WatchJobs, WatchProc, WatchOnce, WatchSystem` 函数
- 作用：创建并返回一个 etcd 的 clientv3.WatchChan，用于监听特定 key 前缀下的所有变化事件（增、删、改）
- 输入：nodeUUID string (除 WatchOnce 外都需要)，用于构造节点专属的监听路径
- 输出：clientv3.WatchChan：一个只读通道，可以从中接收 etcd 的 WatchResponse 事件
- 流程：
    1. 构造Key前缀：使用 fmt.Sprintf 和预定义的 etcdclient 常量（如 etcdclient.KeyEtcdJobProfile）构造要监听的 key 前缀
    2. 发起监听：调用 etcdclient.Watch 并传入 key 前缀和 clientv3.WithPrefix() 选项
    3. 返回通道：直接返回 Watch 函数的结果，供上层逻辑消费

## 6. 辅助函数

#### `JobKey` 函数
- 作用：生成一个在 etcd 中唯一标识某个节点上特定作业的 key
- 输入：nodeUUID string, jobId int
- 输出：`string`：格式化后的 etcd key，如 /crony/jobs/{nodeUUID}/{jobId}

#### `GetJobIDFromKey / GetProcFromKey` 函数
- 作用：从一个 etcd 的 key 字符串中反向解析出结构化信息
- GetJobIDFromKey：从类似 .../jobs/node1/123 的 key 中解析出任务ID 123。
- GetProcFromKey：从类似 .../proc/node1/123/4567 的 key 中解析出 nodeUUID、jobId 和 procId。
- 流程：主要依赖 strings.Split、strings.LastIndex 和 strconv.Atoi 等字符串和类型转换操作。