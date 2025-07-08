config 是 crony项目的配置管理中心, 负责处理与应用程序配置相关的事物. 实现了加载, 解析, 验证和管理应用程序的运行环境和具体配置参数.实现了两大功能:
1. 环境管理: 定义了"testing"和"production"两种运行模式, 并提供了一套机制来从系统变量中读取和验证当前应处的环境
2. 配置文件加载与热更新: 利用 Viper 库, 根据当前环境动态地指定目录结构下自动检查并查找配置文件, 并实现配置热更新功能

---
## 环境管理(Environment)

#### `String()` 方法
- 作用: 让 Environment 类型可以被 fmt 包友好地打印

#### `Production() / Testing()` 方法
- 作用: 提供了方便的方式来获取标准的环境常量

#### `Invalid()` 方法
- 作用: 用于检查一个 Environment 值是否合法

#### `NewGlobalEnvironment()` 函数
- 作用: 从操作系统的环境变量中读取并验证当前应用应该运行在哪个环境下
- 流程: 
    1. 尝试读取名为 ENVIRONMENT 的系统环境变量
    2. 如果环境变量不存在, 返回错误
    3. 如果存在, 将其值转换为 Environment 类型
    4. 调用 Invalid() 检查是否合法
    5. 如果值不合法, 则返回错误
- 输出:
    1. `Environment`: 验证通过的 Environment 值
    2. `error`: 错误信息

## 配置文件加载(Config Loading)

#### `LoadConfig(env, serverName, configFileName string)` 函数
- 作用: 根据环境,服务名,和文件名, 自动查找,加载,解析配置文件, 并启动热更新监控
- 输入: 
    1. `env`: 当前的运行环境, 由`NewGlobalEnvironment()`提供
    2. `serverName`: 服务的名称, 用于构造配置文件路径
    3. `configFileName`: 配置文件的基础名称(不带扩展名)
- 流程: 
    1. 根据输入参数拼接出配置文件的基础目录, 格式为 {服务名}/{命名空间}/{环境}
    2. 自动查找, json,yaml,ini
    3. 创建 Viper 实例
    4. 将找到配置文件路径和类型告知 Viper
    5. 调用 v.ReadInConfig() 将文件读入内存
    6. 调用 v.WatchConfig() 启动一个后台 goroutine 来监控配置文件的变化
    7. 注册一个回调函数 v.OnConfigChange(), 当文件被修改并保存时, 自动调用 v.Unmarshal() 将新的配置内容解析到程序中的配置结构体
    8. 将首次加载的配置内容通过 v.Unmarshal() 解析到一个 models.Config 结构体实例中
    9. 将解析成功的配置实例赋值给包级别的全局变量 _defaultConfig
- 输出:
    1. `*models.Config`: 指向已解析配置的结构体指针
    2. `error`: 如果加载或解析失败, 会panic

#### `GetMysqlDB()` 函数  
- 作用: 一个简单的 getter 函数, 用于在应用的其他地方获取由 LoadConfig 加载并缓存的全局配置实例
- 输出:
    1. `*models.Config`: 指向全局配置实例的指针