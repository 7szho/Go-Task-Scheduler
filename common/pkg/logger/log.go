package logger

import (
	"crony/common/pkg/utils"
	"fmt"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 用于存储全局唯一的 zap 日志记录器实例
var _defaultLogger *zap.Logger

// 初始化全局日志记录器, 接收详细的配置信息, 创建一个高度定制化的 zap.Logger, 并将其赋值给全局的 _defaultLogger
func Init(projectName string, level string, format, prefix, director string, showLine bool, encodeLevel string, stacktraceKey string, logInConsole bool) (logger *zap.Logger) {
	// 1. 检查并创建日志记录
	if ok := utils.Exists(fmt.Sprintf("%s/%s", projectName, director)); !ok { // 判断是否有 Director 文件夹
		fmt.Printf("create %v directory\n", director)
		_ = os.Mkdir(fmt.Sprintf("%s/%s", projectName, director), os.ModePerm)
	}
	// 2. 定义不同日志级别的 LevelEnabler
	debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.DebugLevel
	})
	infoPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.InfoLevel
	})
	warnPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.ErrorLevel
	})
	errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})
	// 3. 根据指定的日志级别, 构建一个或多个 zapcore.Core
	cores := make([]zapcore.Core, 0)
	switch level {
	case "error":
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	case "warn":
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_warn.log", projectName, director), warnPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	case "info":
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_info.log", projectName, director), infoPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_warn.log", projectName, director), warnPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	default:
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_debug.log", projectName, director), debugPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_info.log", projectName, director), infoPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_warn.log", projectName, director), warnPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	}
	// 4. 合并所有的 Core 并创建 Logger
	// zapcore.NewTee 可以将日志同时输出到多个 Core
	// zap.AddCaller() 选项会添加文件名和行号到日志中
	logger = zap.New(zapcore.NewTee(cores[:]...), zap.AddCaller())

	// 根据配置参数, 再次确认是否需要添加调用者信息
	if showLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	// 5. 将创建好的 logger 赋值给全局变量
	_defaultLogger = logger
	return _defaultLogger
}

// getEncoderConfig 创建并返回 zap 的编码器设置
// 编码器配置定义了日志输出的各个字段(如时间, 级别, 消息) 的键名和格式
func getEncoderConfig(prefix, encodeLevel, stacktraceKey string) (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:    "message",                     // 消息字段的 key
		LevelKey:      "level",                       // 级别字段的 key
		TimeKey:       "time",                        // 时间字段的 key
		NameKey:       "logger",                      // 日志记录器名称的 key
		CallerKey:     "caller",                      // 调用者消息的 key (文件名和行号)
		StacktraceKey: stacktraceKey,                 // 堆栈跟踪的 key
		LineEnding:    zapcore.DefaultLineEnding,     // 行结束符 ("\n")
		EncodeLevel:   zapcore.LowercaseLevelEncoder, // 默认级别编码器
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) { // 自定义格式时间
			enc.AppendString(t.Format(prefix + utils.TimeFormatDateV4)) // 使用前缀和预定义的格式时间
		},
		EncodeDuration: zapcore.SecondsDurationEncoder, // Duration 类型字段的编码器
		EncodeCaller:   zapcore.FullCallerEncoder,      // 调用者消息的完整路径编码器
	}
	// 根据传入的参数选择日志级别的编码方式, 支持彩色输出
	switch encodeLevel {
	case "LowercaseLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseLevelEncoder // "info"
	case "LowercaseColorLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder // 彩色的 "info"
	case "CapitalLevelEncoder":
		config.EncodeLevel = zapcore.CapitalLevelEncoder // "INFO"
	case "CapitalColorLevelEncoder":
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder // 彩色的 "INFO"
	default:
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return config
}

// getEncoder 根据指定的格式 (json 或 console) 创建并返回一个 zap 编码器
func getEncoder(prefix, format, encodeLevel, stacktraceKey string) zapcore.Encoder {
	if format == "json" {
		// JSON 格式, 结构化日志, 易于机器解析
		return zapcore.NewJSONEncoder(getEncoderConfig(prefix, encodeLevel, stacktraceKey))
	}
	// Console 格式, 人类易读性刚好
	return zapcore.NewConsoleEncoder(getEncoderConfig(prefix, encodeLevel, stacktraceKey))
}

// getEncoderCore 是一个辅助函数, 用于创建一个完整的 zapcore.Core
// 它将编码器(Encoder), 写入器(WriteSyncer), 和级别控制器(LevelEnabler)组合在一起
func getEncoderCore(logInConsole bool, prefix, format, encodeLevel, stacktraceKey string, fileName string, level zapcore.LevelEnabler) (core zapcore.Core) {
	// 获取日志写入器, 它负责将日志写入文件或控制台
	writer := getWriteSyncer(logInConsole, fileName) // 使用 file-rotatelogs 进行日志分割
	// 创建并返回 Core
	return zapcore.NewCore(getEncoder(prefix, format, encodeLevel, stacktraceKey), writer, level)
}

// getWriteSyncer 创建并返回一个日志写入器(zapcore.WriteSyncer)
// 它使用 lumberjack 实现日志文件的自动分割和归档
func getWriteSyncer(logInConsole bool, file string) zapcore.WriteSyncer {
	// 配置 lumberjack
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file, // 日志文件名
		MaxSize:    10,   // 每个日志的最大大小 (MB)
		MaxBackups: 200,  // 保留旧日志文件的最大数量
		MaxAge:     30,   // 旧日志文件最长保留天数
		Compress:   true, // 是否压缩旧日志文件 (使用gzip)
	}

	if logInConsole {
		// 如果配置为同时输出到控制台, 则使用 NewMultiWriterSyncer
		// 它将日志同时写入标准输出 (os.Stdout) 和 lumberjack(文件)
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	}
	// 否则, 值写入到 lumberjack 文件
	return zapcore.AddSync(lumberJackLogger)
}

// Sync 是对_defaultLogger.Sync() 的一个封装, 用于将缓冲区中的日志刷出(Flush)到目的地
func Sync() error {
	return _defaultLogger.Sync()
}

// Shutdown 是一个用于程序关闭时调用的便捷函数, 用于确保所有挂起的日志都被写入
func Shutdown() {
	_defaultLogger.Sync()
}

// GetLogger 是一个公有的 "getter" 函数, 返回全局的 _defaultLogger 实例
func GetLogger() *zap.Logger {
	return _defaultLogger
}
