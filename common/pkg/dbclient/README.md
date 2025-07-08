实现了一个数据库客户端工具包, 基于流行的 GORM 库, 提供了一套完整且易于配置的 MySQL 数据库连接管理方案.
其核心功能: 
1. 初始化与管理连接池: 通过一个 Init 函数, 建立于 MySQL 数据库的连接, 并配置连接池参数和一个全局共享的数据库实例  
2. 提供灵活的日志系统: 封装了 GORM 的日志功能, 运行通过简单的配置字符串("info", "warn")来动态控制SQL日志的详细程度, 包括慢查询监控和彩色输出
3. 提供辅助工具: 包含一个 CreateDatabase 函数, 用于在应用首次启动时自动创建所需的数据库, 简化了项目的初始化部署流程  

遵循了单例模式, 确保整个应用程序复用同一个数据库连接池, 高效且安全.

--- 

#### `setConfig(logMode string)` 函数  
- 作用: 工具字符串参数, 动态地创建一个 *gorm.Config 对象, 并为其配置好日志记录器
- 输入:
    1. `logMode`: 描述日志级别的字符串
- 流程:
    1. 初始化一个 *gorm.Config 对象
    2. 创建一个默认的, 功能齐全的日志记录器
    3. 设置记录器的输出目标, 慢查询阈值, 默认记录级别, 彩色输出
    4. 动态设置日志级别, 根据不同的 logMode 值, 调用 _default.LogMode() 方法
    5. 将配置好的日志记录器赋值给 config.Logger 字段
- 输出:
    1. `*gorm.Config`: 完整的 *gorm.Config 对象

#### `Init(dsn, logMode string, maxIdleConns, maxOpenConns int)` 函数  
- 作用: 初始化全局数据库连接
- 输入:
    1. `dsn`: 数据库连接字符串(Data Source Name), 包含用户,密码,地址,数据库名等连接信息
    2. `logMode`: GORM 的日志级别
    3. `maxIdleConns`: 连接池中运行存在的最大空闲连接数
    4. `maxOpenConns`: 连接池允许打开的最大连接数
- 流程:
    1. 使用传入的 dsn 配置 GORM 的 MySQL 驱动
    2. 调用 gorm.Open 尝试建立到数据库的连接, 并创建一个 GORM 实例
    3. 如果连接成功, 设置数据库连接池的最大空闲连接数和最大打开连接数
    4. 将创建成功的 GORM 实例赋值给包内的全局变量 _defaultDB, 以便其他函数复用
- 输出:
    1. `*gorm.DB`: 全局共享的 GORM 数据库实例的指针
    2. `error`: 错误信息

#### `GetMysqlDB()` 函数  
- 作用: 获取已经初始化好的全局数据库实例
- 流程:
    1. 检查全局变量 _defaultDB 是否为 nil, 即是否已经调用 Init 函数进行过初始化
    2. 如果尚未初始化, 就记录一条错误日志并返回 nil
    3. 如果已经初始化, 就返回这个全局的 _defaultDB 实例
- 输出:
    1. `*gorm.DB`: 全局共享的 GORM 数据库实例的指针

#### `CreateDatabase(dsn string, driver string, createSql string)` 函数  
- 作用: 一个辅助函数, 用于在数据库服务器上创建一个新的数据库
- 输入:
    1. `dsn`: 数据库连接字符串(Data Source Name), 包含用户,密码,地址等连接信息
    2. `driver`: 数据库驱动的名称
    3. `createSql`: 要执行的 SQL 语句
- 流程:
    1. 使用 Go 原生的 database/sql 包打开一个到数据库服务器的连接
    2. 使用 db.Ping() 检查与服务器的网络连通性
    3. 如果连接成功, 执行传入的 createSql 语句来创建数据库
    4. 使用 defer 确保函数结束时关闭这个临时的数据库连接
- 输出:
    1. `error`: 错误信息