package conf

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
)

type GateWayConfig struct {
	ServiceId     string
	ServiceName   string `default:"gateway"`
	Listen        string `default:":8000"`
	PublicAddress string
	PublicPort    int `default:"8000"`
	Tags          []string
	ConsulURL     string
	MonitorPort   int `default:"8001"`
	AppSecret     string
	LogLevel      string `default:"INFO"`
}

func (c GateWayConfig) String() string {
	bts, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(bts)
}

// InitGateWayConfig 初始化网关配置选项，从配置文件file读取
func InitGateWayConfig(file string) (*GateWayConfig, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config file not found: %v", err)
	}

	var config GateWayConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
