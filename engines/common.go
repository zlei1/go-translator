package engines

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	YoudaoAppKey    string `mapstructure:"youdao_app_key"`
	YoudaoAppSecret string `mapstructure:"youdao_app_secret"`
	BaiduAppKey    string `mapstructure:"baidu_app_key"`
	BaiduAppSecret string `mapstructure:"baidu_app_secret"`
}

func readConfigFile() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath("/etc/go-translator/")
	viper.AddConfigPath("./config/")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}

	config := new(Config)
	err = viper.Unmarshal(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}

	return config
}
