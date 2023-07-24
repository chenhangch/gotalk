package conf

import (
	"fmt"
	"github.com/spf13/viper"
)

type LogicServerConfig struct {
	ServerId      string
	NameSpace     string
	Listen        string
	PublicAddress string
	PublicPort    int
	Tags          []string
	ConsulRUL     string
	RedisAddr     string
	RpcURL        string
}

// InitLogicConfig initial logicServer configuration
func InitLogicConfig(file string) (*LogicServerConfig, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config file %v: %v", file, err)
	}
	var config LogicServerConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
