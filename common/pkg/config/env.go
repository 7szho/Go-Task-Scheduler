package config

import (
	"errors"
	"os"
)

// use export ENVIRONMENTS=testing set global environment
const (
	EnvTesting    = Environment("testing")    // 测试环境
	EnvProduction = Environment("production") // 生产环境
)

type Environment string

// String 方法让 Environment 类型实现了 fmt.Stringer 接口
// 当使用fmt包函数打印 Environment 类型的变量时, 会调用该方法获取其字符串表示
func (env *Environment) String() string {
	return string(*env)
}

// Production 方法返回一个表示生产环境的 Environment 变量
func (env *Environment) Production() Environment {
	return EnvProduction
}

// Testing 方法返回一个表示测试环境的 Environment 变量
func (env *Environment) Testing() Environment {
	return EnvTesting
}

// Invalid 方法用于检查一个 Environment 值是否是无效的
func (env Environment) Invalid() bool {
	return env != EnvTesting && env != EnvProduction
}

// NewGlobalEnvironment 读取全局配置的环境变量
func NewGlobalEnvironment() (Environment, error) {
	// 读取环境变量
	environment, ok := os.LookupEnv("ENVIRONMENT")
	// 检查环境变量是否存在
	if !ok {
		return "", errors.New("system environment:ENVIRONMENT not found")
	}

	// 将从环境变量中读取的字符串值强制转换为 Environment 类型
	env := Environment(environment)
	// 检查是否无效
	if env != EnvTesting && env != EnvProduction {
		return "", errors.New("environment not support, must be production, development")
	}

	// 如果通过所有检查, 返回解析出的环境env
	return env, nil
}
