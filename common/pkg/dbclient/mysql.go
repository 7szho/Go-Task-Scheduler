package dbclient

import (
	"crony/common/pkg/logger"
	"database/sql"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var _defaultDB *gorm.DB

// Init 函数负责初始化数据库连接
//
// @param dsn string: 数据库源名称, 包含了连接所需的所有信息
// @param logMode string: GORM 的日志模式
// @param maxIdleConns int: 连接池中最大空闲连接数
// @param maxOpenConns int: 连接池中最大打开连接数
// @return *gorm.DB: 成功时返回 GORM 数据库实例的指针
func Init(dsn, logMode string, maxIdleConns, maxOpenConns int) (*gorm.DB, error) {
	// 配置 GORM 的 Mysql 驱动
	mysqlConfig := mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		SkipInitializeWithVersion: false,
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), setConfig(logMode)); err != nil {
		return nil, err
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		_defaultDB = db
		return db, nil
	}
}

// getter 函数, 返回已初始化的数据库实例
func GetMysqlDB() *gorm.DB {
	if _defaultDB == nil {
		logger.GetLogger().Error("mysql database is not initialized")
		return nil
	}
	return _defaultDB
}

// 辅助函数, 用于执行创建数据库的SQL语句
//
// @param dsn string: 连接数据库服务器的 DSN
// @param driver string: 数据库驱动名
// @param createSql string: 要执行了 `CREATE DATABASE...` 语句
func CreateDatabase(dsn string, driver string, createSql string) error {
	// sql.Open 只是验证参数并返回一个 *sql.DB 连接池对象, 并不会直接建立连接
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}

	// 使用 defer 确保在函数结束时关闭数据库连接池, 释放资源
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(db)

	// db.Ping() 会尝试与数据库建立一个连接, 并检查连接是否有效
	// 这是实际发生网络通信的地方
	if err = db.Ping(); err != nil {
		return err
	}
	// 执行传入的 SQL 语句
	_, err = db.Exec(createSql)

	// 返回错误
	return err
}
