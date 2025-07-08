package dbclient

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type writer struct {
	logger.Writer
}

func newWriter(w logger.Writer) *writer {
	return &writer{Writer: w}
}

// Config Custom Gorm
// setConfig 函数根据传入的日志模式字符串, 创建一个自定义的 GORM 配置对象
func setConfig(logMode string) *gorm.Config {
	// DisableForeignKeyConstraintWhenMigrating: true 表示自动迁移数据库表结构时, 不创建外键约束, 避免循环依赖问题
	config := &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}
	_default := logger.New(newWriter(log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Warn,
		Colorful:      true,
	})
	// 使用 switch 语句, 根据传入的 logMode 字符串动态设置最终的日记记录器
	switch logMode {
	case "silent", "Silent":
		// Silent 模式: 不记录任何日志
		config.Logger = _default.LogMode(logger.Silent)
	case "error", "Error":
		// Error 模式: 只记录错误日志
		config.Logger = _default.LogMode(logger.Error)
	case "warn", "Warn":
		// Warn 模式: 记录警告和错误日志
		config.Logger = _default.LogMode(logger.Warn)
	case "info", "Info":
		// Info 模式: 记录所有级别日志
		config.Logger = _default.LogMode(logger.Info)
	default:
		// logMode 字符串没有匹配上, 默认使用 Info 模式
		config.Logger = _default.LogMode(logger.Info)
	}
	// 返回最终配置好的 gorm.Config 对象
	return config
}
