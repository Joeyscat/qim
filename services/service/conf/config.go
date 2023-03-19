package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/logger"
	"github.com/kataras/iris/v12/middleware/accesslog"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Server struct {
}

type Config struct {
	ServerID      string
	NodeID        int64
	Listen        string `default:":8080"`
	PublicAddress string
	PublicPort    uint16 `default:"8080"`
	Tags          []string
	EtcdEndpoints string
	RedisAddrs    string
	Driver        string `default:"mysql"`
	BaseDB        string
	MessageDB     string
	LogLevel      string `default:"debug"`
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

	if err := viper.ReadInConfig(); err != nil {
		logger.L.Warn(err.Error())
	} else {
		if err := viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	}

	err := envconfig.Process("qim", &config)
	if err != nil {
		return nil, err
	}

	if config.ServerID != "" {
		localIP := qim.GetLocalIP()
		config.ServerID = fmt.Sprintf("royal_%s", strings.ReplaceAll(localIP, ".", ""))
		arr := strings.Split(localIP, ".")
		if len(arr) == 4 {
			suffix, _ := strconv.Atoi(arr[3])
			config.NodeID = int64(suffix)
		}
	}
	if config.PublicAddress == "" {
		config.PublicAddress = qim.GetLocalIP()
	}
	logger.L.Debug("load config finished", zap.String("config", config.String()))

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

func MakeAccessLog() *accesslog.AccessLog {
	// create a new access log middleware.
	ac := accesslog.File("./access.log")
	// remove this if you don't want to log to the console.
	ac.AddOutput(os.Stdout)

	// the default configuration:
	ac.Delim = '|'
	ac.TimeFormat = "2006-01-02 15:04:05"
	ac.Async = false
	ac.IP = true
	ac.BytesReceivedBody = true
	ac.BytesSentBody = true
	ac.BytesReceived = false
	ac.BytesSent = false
	ac.BodyMinify = true
	ac.RequestBody = true
	ac.ResponseBody = false
	ac.KeepMultiLineError = true
	ac.PanicLog = accesslog.LogHandler

	return ac
}
