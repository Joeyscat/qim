package conf

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/logger"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	ServerID        string
	ServerName      string `default:"wgateway"`
	Listen          string `default:":8000"`
	PublicAddress   string
	PublicPort      uint16 `default:"8000"`
	Tags            []string
	Domain          string
	EtcdEndpoints   string
	MonitorPort     uint16 `default:"8001"`
	AppSecret       string
	LogLevel        string `default:"debug"`
	MessageGPool    int    `default:"10000"`
	ConnectionGPool int    `default:"15000"`
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
		config.ServerID = fmt.Sprintf("gate_%s", strings.ReplaceAll(localIP, ".", ""))
	}
	if config.PublicAddress == "" {
		config.PublicAddress = qim.GetLocalIP()
	}
	logger.L.Debug("load config finished", zap.Any("config", config))

	return &config, nil
}
