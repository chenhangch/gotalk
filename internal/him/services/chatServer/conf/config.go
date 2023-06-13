package conf

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/spf13/viper"
	"time"
)

type ChatServerConfig struct {
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

// InitChatConfig initial chatServer configuration
func InitChatConfig(file string) (*ChatServerConfig, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config file %v: %v", file, err)
	}
	var config ChatServerConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func InitRedis(addr string, pass string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	})

	_, err := rdb.Ping().Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}

func InitFailoverRedis(masterName string, sentinelAddrs []string, pass string, timeout time.Duration) (*redis.Client, error) {
	redisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      pass,
		DialTimeout:   time.Second * 5,
		ReadTimeout:   timeout,
		WriteTimeout:  timeout,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}
