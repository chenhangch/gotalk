package conf

import (
	"encoding/json"
	"github.com/spf13/viper"
)

// ServiceConfig Configuration
type ServiceConfig struct {
	ServiceId     string
	NodeID        int64
	Listen        string `default:":8080"`
	PublicAddress string
	PublicPort    int `default:"8080"`
	Tags          []string
	ConsulURL     string
	BaseDb        string
	MessageDb     string
	LogLevel      string `default:"INFO"`
}

func (c ServiceConfig) String() string {
	bts, _ := json.Marshal(c)
	return string(bts)
}

func InitServiceConfig(file string) (*ServiceConfig, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")

	var config ServiceConfig
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	} else {
		if err := viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	}
	// TODO: 读取环境变量中的配置

	return &config, nil
}
