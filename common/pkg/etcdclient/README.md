etcdclient 包是 crony 项目与 etcd 基础交互的唯一入口和高级封装层, 集成了连接管理,服务注册与发现, 分布式锁等多种功能. 核心功能为:
1. 定义数据模型与常量: 以常量形式统一管理了所有存储再etcd中的键(key)的格式和前缀
2. 封装核心API: 对 etcd 官方客户端的常用操作进行二次封装. 这些封装不仅简化了调用, 还统一加入了超时控制,初始化检查,日志记录,和错误处理, 并实现了分布式锁等高级功能
3. 实现服务注册: 定义了 ServerReg 结构体, 用于实现服务自动注册与心跳维持的逻辑
4. 定义通用接口: 定义了一个 Watcher 接口, 为所有需要"监视"etcd中数据变化的组件提供了一个抽象规范

--- 

## `const.go`  

集中定义了所有存储在 etcd 中的键(key)的结构和前缀模板   
keyEtcdProfile: 根路径 /crony/。  
KeyEtcdNodeProfile 和 KeyEtcdNode: 用于节点注册的键，如 /crony/node/{node_uuid}。  
KeyEtcdProcProfile 系列: 用于记录正在执行的任务进程的键，如 /crony/proc/{node_uuid}/{job_id}/{pid}。  
KeyEtcdJobProfile 和 KeyEtcdJob: 用于存储任务定义的键，如 /crony/job/{node_uuid}/{job_id}。  
KeyEtcdOnceProfile 和 KeyEtcdOnce: 用于一次性任务的键。  
KeyEtcdLockProfile 和 KeyEtcdLock: 用于分布式锁的键。  
KeyEtcdSystemProfile 系列: 用于系统控制命令的键。  

## `etcdclient.go`

#### `Init()` 函数
- 作用: 初始化全局唯一的 etcd 客户端实例 _defaultEtcd
- 输入: 
    1. `endpoints`: etcd 集群的节点地址列表
    2. `dialTimeout`: 连接超时时间
    3. `reqTimeout`: 请求超时时间
- 流程: 
    1. 创建一个 clientv3.Config 对象, 将输入的参数填充进去
    2. 调用 clientv3.New() 使用该配置创建官方的 etcd 客户端实例
    3. 如果创建失败, 打印错误日志并返回错误
    4. 如果成功, 将官方客户端实例和一个 reqTimeout 包装进自定义的 Client 结构体中
    5. 将这个自定义的 Client 实例赋值给包级全局变量 _defaultEtcd
- 输出:
    1. `*Client`:  指向成功初始化的自定义客户端实例的指针
    2. `error`: 错误信息

#### `GetEtchClient()` 函数  
- 作用: 获取已经初始化好的 etcd 实例
- 流程:
    1. 检查全局变量 _defaultEtcd 是否为 nil, 即是否已经调用 Init 函数进行过初始化
    2. 如果尚未初始化, 就记录一条错误日志并返回 nil
    3. 如果已经初始化, 就返回这个全局的 _defaultEtcd 实例
- 输出:
    1. `*Client`: 全局共享的 etcd 客户端实例的指针

#### `Put(key, val string, opts ...clientv3.OpOption)` 函数
- 作用: 封装了标准的 etcd Put 操作, 用于创建或更新一个键值对
- 输入: 
    1. `key`: 要操作的键
    2. `val`: 要存入的值
    3. `opts`: 可选的 Put 操作选项(如 WithLease, WithPrevKV等)
- 流程: 
    1. 检查客户端是否已初始化
    2. 调用 NewEtcdTimeoutContext() 创建一个带超时时间的 context
    3. 使用 defer cancel() 确保 context 资源被释放
    4. 调用 _defaultEtcd.Client.Put() 方法执行实际的写入操作
- 输出:
    1. `*clientv3.PutResponse`: etcd 返回的响应
    2. `error`: 错误信息

#### `PutWithTtl(key, val string, ttl int64)` 函数
- 作用: 创建一个带租约（Time-To-Live）的键值对. 当租约到期后, 该键值对会自动从 etcd 中删除
- 输入: 
    1. `key`: 要操作的键
    2. `val`: 要存入的值
    3. `ttl`: 租约的生命周期(秒)
- 流程: 
    1. 检查客户端是否已初始化
    2. 调用 Grant(ttl) 申请一个新的租约, 并获取租约 ID
    3. 如果申请租约失败, 返回错误
    4. 调用 Put() 方法, 并在可选参数中传入 clientv3.WithLease(leaseRsp.ID), 将键值对与租约绑定
- 输出:
    1. `*clientv3.PutResponse`: Put 操作的响应
    2. `error`: 错误信息

#### `PutWithModRev(key, val string, rev int64)` 函数
- 作用: 实现基于版本号的“比较并交换”（CAS）原子操作. 只有当服务器上该 key 的 ModRevision 与传入的 rev 相同时, 更新才会成功
- 输入: 
    1. `key`: 要操作的键
    2. `val`: 要存入的值
    3. `rev`: 期望的 ModRevision 版本号
- 流程: 
    1. 检查客户端是否已初始化
    2. 如果版本号为0, 退化为普通的 Put 操作
    3. 创建一个 etcd 事务 (Txn)
    4. 在事物的 If 部分, 设置条件为 `clientv3.Compare(clientv3.ModRevision(key), "=", rev)`
    5. 在 Then 部分, 设置操作为 clientv3.OpPut(key, val)
    6. 提交事务
    7. 检查事务响应的 Succeeded 字段. 如果为 false, 表示 If 条件不满足, 返回错误
    8. 如果成功, 从事务响应中解析出 PutResponse 并返回
- 输出:
    1. `*clientv3.PutResponse`: Put 操作的响应
    2. `error`: 错误信息

#### `Get(key string, opts ...clientv3.OpOption)` 函数
- 作用: 封装了标准的 etcd Get 操作, 用于根据键获取值
- 输入: 
    1. `key`: 要查询的键
    2. `opts`: 可选的操作选项
- 流程: 
    1. 检查客户端是否已初始化
    2. 调用 NewEtcdTimeoutContext() 创建一个带超时时间的 context
    3. 使用 defer cancel() 确保 context 资源被释放
    4. 调用 _defaultEtcd.Client.Get() 方法执行实际的查询操作
- 输出:
    1. `*clientv3.GetResponse`: etcd 返回的响应
    2. `error`: 错误信息

#### `Delete(key string, opts ...clientv3.OpOption)` 函数
- 作用: 封装了标准的 etcd Delete 操作,用于删除一个或多个键
- 输入: 
    1. `key`: 要查询的键
    2. `opts`: 可选的操作选项
- 流程: 
    1. 检查客户端是否已初始化
    2. 调用 NewEtcdTimeoutContext() 创建一个带超时时间的 context
    3. 使用 defer cancel() 确保 context 资源被释放
    4. 调用 _defaultEtcd.Client.Delete() 方法执行实际的删除操作
- 输出:
    1. `*clientv3.DeleteResponse`: etcd 返回的响应
    2. `error`: 错误信息

#### `Watch(key string, opts ...clientv3.OpOption)` 函数
- 作用: 封装了标准的 etcd Watch 操作, 用于监视一个键或一个前缀的变化
- 输入: 
    1. `key`: 要监视的键或前缀
    2. `opts`: 可选的操作选项
- 流程: 
    1. 直接调用 _defaultEtcd.Client.Watch()
    2. 使用了 context.Backgroud(), 意味着 Watcher的生命不受请求超时影响, 需要调用方自己管理
- 输出:
    1. `clientv3.WatchChan`: 一个只读通道, 可以从中接收键变化的事件

#### `Grant(ttl int64)` 函数
- 作用: 封装了申请租约的操作
- 输入: 
    1. `ttl`: 租约的生命周期
- 流程: 
    1. 检查客户端是否已初始化
    2. 调用 NewEtcdTimeoutContext() 创建一个带超时时间的 context
    3. 使用 defer cancel() 确保 context 资源被释放
    4. 调用 _defaultEtcd.Client.Grant() 方法执行实际的申请租约操作
- 输出:
    1. `*clientv3.LeaseGrantResponse`: etcd 返回的响应
    2. `error`: 错误信息

#### `Revoke(id clientv3.LeaseID)` 函数
- 作用: 封装了撤销租约的操作. 撤销租约后, 与该租约关联的所有键值对都会被删除
- 输入: 
    1. `id`: 要撤销的租约 ID
- 流程: 
    1. 检查客户端是否已初始化
    2. 调用 NewEtcdTimeoutContext() 创建一个带超时时间的 context
    3. 使用 defer cancel() 确保 context 资源被释放
    4. 调用 _defaultEtcd.Client.Revoke() 方法执行实际的撤销租约操作
- 输出:
    1. `*clientv3.LeaseRevokeResponse`: etcd 返回的响应
    2. `error`: 错误信息

#### `GetLock(key string, id clientv3.LeaseID)` 函数
- 作用: 尝试获取一个分布式锁. 这是一个非阻塞操作
- 输入: 
    1. `key`: 锁的名称(不含前缀)
    2. `id`: 一个已经申请好的租约 ID, 用于绑定锁
- 流程: 
    1. 使用 KeyEtcdLock 常量和输入 key 拼接成完整的锁路径
    2. 创建一个 etcd 事务
    3. If 条件: 检查锁路径的 CreateRevision 是否为0(即该键不存在)
    4. Then 操作: 如果键不存在, 就创建它, 并将值设为空字符串, 同时绑定租约 id
    5. 提交事务
- 输出:
    1. `bool`: true表示成功获得锁, false表示已被他人持有
    2. `error`: 错误信息

#### `DelLock(key string, id clientv3.LeaseID)` 函数
- 作用: 是否一个分布式锁
- 输入: 
    1. `key`: 锁的名称(不含前缀)
- 流程: 
    1. 调用 Delete() 函数删除对应的锁键
- 输出:
    1. `error`: 错误信息

#### `IsValidAsKeyPath(s string)` 函数
- 作用: 这是一个验证工具函数, 用于检查给定的字符串 s 是否可以作为一个安全的 etcd 键的一部分, 而不是一个包含层级机构的完整路径
- 输入: 
    1. `s`: 需要被检查的字符串
- 流程: 
    1. 调用 Go 标准库的 strings.IndexAny(s, "/\\") 方法
    2. 这个方法会在字符串 s 中查找是否存在任何一个在第二个参数 ("/\\") 中出现的字符, 即 / 或 \
    3. 如果找到了, IndexAny 会返回该字符在 s 中的索引
    4. 如果没找到, IndexAny 会返回 -1
    5. 函数最终返回 strings.IndexAny(s, "/\\") == -1 的布尔结果
- 输出:
    1. `bool`: true表示字符串有效, false表示字符串无效

#### `(c *etcdTimeoutContext) Err()` 方法  
- 作用: 重写 (Override) 了内嵌 context.Context 的 Err() 方法. 这是实现功能增强的关键. 当 context 因为超时或被取消而结束时, 外部代码会调用 Err() 方法来获取原因
- 输入: 
    1. `c`: 方法的接收者
- 流程:
    1. 首先, 调用 c.Context.Err(), 即调用被内嵌的标准 context 的原始 Err() 方法, 获取其返回的错误
    2. 检查这个错误是否 正好 是context.DeadlineExceeded(超时错误)
    3. 如果是超时错误, 就创建一个新的,信息更丰富的字符串. 这个新字符串包含了原始的超时错误信息 (%s), 并附加上了当前 etcd 集群的地址列表 (%v)
    4. 如果不是超时错误, 则直接返回原始错误
    5. 返回最终处理过的 err
- 输出:
    1. `error`:  一个标准的错误消息. 如果是超时, 会是增强后的错误; 否则是原始错误

#### `NewEtcdTimeoutContext()` 函数  
- 作用: 一个工厂函数, 负责创建并正确初始化上面定义的 etcdTimeoutContext. 它将复杂的创建过程封装起来, 让调用者可以像使用标准的 context.WithTimeout 一样方便地获取这个增强版的 context
- 流程:
    1. 从全局客户端实例 _defaultEtcd 中获取预设的请求超时时间 reqTimeout
    2. 调用 context.WithTimeout(context.Background(), ...) 创建一个具有超时功能的标准 context (ctx) 和一个 cancel 函数
    3. 创建一个空的 etcdTimeoutContext 实例 etcdCtx
    4. 将刚刚创建的标准 ctx 赋值给 etcdCtx 的内嵌字段 Context
    5. 从全局配置 config.GetConfigModels() 中获取 etcd 的端点地址列表，并赋值给 etcdCtx 的扩展字段 etcdEndpoints
    6. 将这个完整配置好的 etcdCtx 和原始的 cancel 函数返回
- 输出:
    1. `context.Context`: 一个接口类型, 实际返回的是 *etcdTimeoutContext 指针, 但调用者可以像使用任何标准 context 一样使用它
    2. `context.CancelFunc`: 与创建的 context 配套的取消函数, 调用者必须在适当的时候调用它以释放资源

## `registry.go`  

#### `NewServerReg(ttl int64)` 函数  
- 作用: 一个工厂函数, 创建一个 ServerReg 服务注册器实例
- 输入: 
    1. ttl: 服务注册时使用的租约 TTL(秒)
- 流程:
    1. 创建一个ServerReg结构体实例
    2. 将全局的 _defaultEtcd 客户端赋值给它的 Client 字段
    3. 将输入的 ttl 赋值给它的 Ttl 字段
    4. 初始化 stop channel
- 输出:
    1. `*ServerReg`: 一个已部分初始化的服务注册器指针

## `watcher.go`  

#### `Watcher()` 接口
- 作用: 定义了一个通用的"监视器"主键必须满足的行为契约
- 方法:
    1. Wathc() error: 启动监视逻辑
    2. Close() error: 停止监视并清理资源

任何需要监听 etcd 变化的模块(如任务管理器,节点管理器), 只要实现了这个接口, 就可以被上层统一管理, 实现程序的解耦.