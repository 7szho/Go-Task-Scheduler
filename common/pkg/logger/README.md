logger 包提供了一个高度可配置,高性能且生成就绪的日志服务. 基于 go.uber.org/zap 库构建, 并集成了 github.com/natefinch/lumberjack 来实现日志文件的自动分割,归档,和压缩. 该包遵循单例模式, 通过 Init 函数根据配置初始化一个全局唯一的日志记录器实例. 这个实例随后可以通过 GetLogger() 在项目的任何地方被获取或使用. 其核心功能包括:
1. 分级日志记录: 支持 DEBUG, INFO, WARN, ERROR 等不同的级别日志
2. 多目的地输出: 能够同时将日志输出到文件和控制台
3. 日志轮转: 自动按文件大小,保存时间,和数量对日志进行管理, 防止单个日志文件无限增长
4. 高度可定制: 日志的几乎所有方面, 如时间格式,级别显示方式,是否显示代码行号等, 都可以通过 Init 函数的参数进行配置

--- 

#### `Init(projectName string, level string, format, prefix, director string, showLine bool, encodeLevel string, stacktraceKey string, logInConsole bool)` 函数  
- 作用: 创建一个高度定制化的 zap.Logger, 并将其赋值给全局的 _defaultLogger 供其他地方使用
- 输入:
    1. `projectName`: 项目名称, 用于构建日志文件路径
    2. `level`: 全局日志级别 ("debug", "info", "warn", "error")
    3. `format`: 日志格式 ("json" 或 "console")
    4. `prefix`: 添加在日志时间戳前的前缀
    5. `director`: 存放日志文件的目录名
    6. `showLine`: 是否在日志中显示调用者的文件名和行号
    7. `encodeLevel`: 日志级别的显示方式
    8. `stacktraceKey`: 在日志中堆栈跟踪信息的字段名
    9. `logInConsole`: 是否同时将日志打印到控制台
- 流程:
    1. 检查并创建日志文件存放的目录
    2. 定义不同日志级别的 LevelEnabler, 它们是 zap 的过滤器, 决定了一个 Core 只处理哪个级别的日志
    3. 根据输入的 level 参数, 使用 switch 语句构建一个或多个 zapcore.Core
    4. 使用 zapcore.NewTee(cores...) 将所有创建的 Core 合并. NewTee 的作用是将一条日志消息同时发送到所有这些 Core 中, 每个 Core 再根据自己的 LevelEnabler 决定是否处理
    5. 使用 zap.New() 基于合并后的 Tee Core 创建最终的 zip.Logger 实例, 并默认添加 zap.AddCaller() 选项
    6. 根据 showLine 参数再次确认是否需要添加调用者信息
    7. 将最终创建的 logger 赋值给全局变量 _defaultLogger
- 输出:
    1. `logger`: 返回创建好的,可供使用的 zap 日志记录器实例

#### `getEncoderConfig(prefix, encodeLevel, stacktraceKey string)` 函数  
- 作用: 一个辅助函数, 用于创建和配置 zapcore.EncoderConfig. 这个配置对象定义了日志条目中各个字段的格式和键名
- 输入:
    1. `prefix`: 时间戳前缀
    2. `encodeLevel`: 日志级别的编码方式
    3. `stacktraceKey`: 堆栈跟踪的键名
- 流程:
    1. 创建一个 zapcore.EncoderConfig 结构体
    2. 设置 MessageKey, LevelKey, TimeKey 等标准字段的键名
    3. 定义一个自定义的时间编码器 EncodeTime, 它使用传入的 prefix 和预定义的格式来格式化时间
    4. 根据 encodeLevel 参数, 通过 switch 语句选择一个合适的级别编码器
- 输出:
    1. `config`: 返回一个完全配置好的编码器设置对象

#### `getEncoder(prefix, format, encodeLevel, stacktraceKey string)` 函数  
- 作用: 一个辅助函数, 根据指定的格式创建并返回一个具体的 zap 编码器实例. 编码器负责将日志条目序列化成最终的字符串格式
- 输入:
    1. `prefix`: 时间戳前缀
    2. `format`: 日志格式 ("json" 或 "console")
    3. `encodeLevel`: 日志级别的编码方式
    4. `stacktraceKey`: 堆栈跟踪的键名
- 流程:
    1. 调用 getEncoderConfig() 获得基础的编码器配置
    2. 如果 format 是 "json", 则返回 zapcore.NewJSONEncoder
    3. 否则, 默认返回 zapcore.NewConsoleEncoder
- 输出:
    1. `zapcore.Encode`: 返回一个实现了 Encoder 接口的具体实例

#### `getEncoderCore(logInConsole bool, prefix, format, encodeLevel, stacktraceKey string, fileName string, level zapcore.LevelEnabler)` 函数  
- 作用: 一个核心的内部组装函数. 它将编码器(Encoder),写入器(WriteSyncer),和级别控制器(LevelEnabler)这三个 zap 的核心组件组合在一起
- 输入: 包含了创建 Encoder 和 WriteSyncer 所需的所有配置, 以及一个 LevelEnabler 过滤器
- 流程:
    1. 调用 getWriteSycer() 获取日志的写入目的地
    2. 调用 getEncoder() 获取日志的编码格式
    3. 使用 zapcore.NewCore() 将上面两步的结果与传入的 level 过滤器组合起来
- 输出:
    1. `core`: 返回一个组装好的,功能齐全的 zapcore.Core 实例

#### `getWriteSyncer(logInConsole bool, file string)` 函数  
- 作用: 创建并返回一个日志的写入目的地. 它使用 lumberjack 库来实现日志文件的自动分割,归档,和压缩
- 输入:
    1. `logInConsole`: 是否也输出到控制台
    2. `file`: 日志文件的完整路径
- 流程:
    1. 配置一个 lumberjack.Logger 实例, 设置好日志文件名,最大大小,最大备份数,最大保存天数,和是否压缩
    2. 判断 logInConsole 标志
    3. 如果为 true, 使用 zapcore.NewMultiWriteSyncer 将 os.Stdout (标准输出) 和 lumberjack 文件写入器合并, 实现同时向两处写入
    4. 如果为 false, 则只返回 lumberjack 文件写入器
- 输出:
    1. `zapcore.WriteSyncer`: 返回一个实现了WriteSyncer 接口的写入器实例

#### `GetLogger() *zap.Logger` 函数  
- 作用: 一个 getter 函数, 遵循单例模式, 返回 Init 函数中创建并存储的全局 _defaultLogger 实例
- 输出: `zap.Logger`: 全局唯一的日志记录器实例

#### `Sync` 和 `Shutdown` 函数  
- 作用: 这两个是便捷函数, 用于确保在程序退出前, 所有在内存缓冲区中的日志都被刷出 (Flush) 到最终的目的地 (文件或控制台). Zap 为了性能，会先将日志写入内存缓冲区.
- 流程:
    - `Sync()`: 直接调用 _defaultLogger.Sync() 并返回错误
    - `Shutdown()`: 只是调用 _defaultLogger.Sync(), 忽略其错误
- 输出:
    1. `Sync()`: 返回一个 error
