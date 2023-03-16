package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/logger"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Server struct {
}

type Config struct {
	ServerID        string
	Listen          string `default:":8005"`
	MonitorPort     uint16 `default:"8006"`
	PublicAddress   string
	PublicPort      uint16 `default:"8005"`
	Tags            []string
	Zone            string `default:"zone_03"`
	EtcdEndpoints   string
	RedisAddrs      string
	RoyalURL        string
	LogLevel        string `default:"debug"`
	MessageGPool    int    `default:"5000"`
	ConnectionGPool int    `default:"500"`
}

func (c Config) String() string {
	bts, _ := json.Marshal(c)
	return string(bts)
}

func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")

	var config Config

	err := envconfig.Process("qim", &config)
	if err != nil {
		return nil, err
	}

	if err := viper.ReadInConfig(); err != nil {
		logger.L.Warn(err.Error())
	} else {
		if err := viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	}

	if config.ServerID != "" {
		localIP := qim.GetLocalIP()
		config.ServerID = fmt.Sprintf("server_%s", strings.ReplaceAll(localIP, ".", ""))
	}
	if config.PublicAddress == "" {
		config.PublicAddress = qim.GetLocalIP()
	}
	logger.L.Debug("load config finished", zap.Any("config", config))

	return &config, nil
}

func InitRedis(addr string, pass string) (*redis.Client, error) {
	redisdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	})

	_, err := redisdb.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}

	return redisdb, nil
}

func InitFailoverRedis(masterName, password string, sentinelAddrs []string, timeout time.Duration) (*redis.Client, error) {
	redisdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
		DialTimeout:   time.Second * 5,
		ReadTimeout:   timeout,
		WriteTimeout:  timeout,
	})
	_, err := redisdb.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}
	return redisdb, nil
}
