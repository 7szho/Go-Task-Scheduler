package config

import (
	"crony/common/models"
	"crony/common/pkg/utils"
	"fmt"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// 定义支持的配置文件扩展名常量
const (
	ExtensionJson = ".json"
	ExtensionYaml = ".yaml"
	ExtensionInI  = ".ini"

	// 定义了配置文件夹的固定名称
	NameSpace = "conf"
)

var (
	// Automatic loading sequence of local Config
	autoLoadLocalConfigs = []string{
		ExtensionJson,
		ExtensionYaml,
		ExtensionInI,
	}
)

var _defaultConfig *models.Config

// LoadConfig 函数加载配置
// 它根据env, serverName, 和configFileName来查找并加载配置
func LoadConfig(env, serverName, configFileName string) (*models.Config, error) {
	var c models.Config // 用于存储解析后的配置
	var confPath string // 用于存放最终找到的配置文件路径

	// 1. 构造配置文件的基础目录路径
	// 格式为: {服务名}/{命名空间}/{环境}
	dir := fmt.Sprintf("%s/%s/%s", serverName, NameSpace, env)

	// 2. 自动查找配置文件
	// 遍历预设的扩展名列表
	for _, registerExt := range autoLoadLocalConfigs {
		// 使用 path.Join 拼接完整的文件路径
		confPath = path.Join(dir, configFileName+registerExt)
		if utils.Exists(confPath) {
			break
		}
	}
	// 打印出最终决定使用的配置文件路径, 方便调试
	fmt.Println("the path to the configuration file you are using is: ", confPath)

	// 3. 使用 Viper 加载和解析配置
	v := viper.New()          // 创建一个新的 Viper 实例, 避免使用全局单例
	v.SetConfigFile(confPath) // 明确告诉 Viper, 要加载哪个文件

	ext := utils.Ext(confPath)
	v.SetConfigType(ext) // 告诉 Viper 配置文件的类型, 以便它使用正确的解析器

	// 读取配置文件内容到 Viper 实例中
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("FATAL ERROR CONFIG FILE: %s", err))
	}

	// 4. 启用配置热加载
	v.WatchConfig() // 启动对配置文件的监控
	// 注册一个回调函数, 当配置文件发生变化时, 这个函数会被调用
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed: ", e.Name)
		if err := v.Unmarshal(&c); err != nil {
			fmt.Println(err)
		}
	})

	// 5. 将初次加载的配置解析到结构体中
	if err := v.Unmarshal(&c); err != nil {
		panic(fmt.Errorf("FATAL ERROR CONFIG FILE: %s", err))
	}
	// 打印加载后的配置内容, 方便调试
	fmt.Printf("load config is: %#v \n", c)

	// 6. 将加载成功的配置实例赋值给包级别的全局变量
	_defaultConfig = &c

	// 返回配置实例的指针和nil错误
	return &c, nil
}

// getter函数, 返回缓存的全局配置实例
func GetConfigModels() *models.Config {
	return _defaultConfig
}
