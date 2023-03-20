package conf

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/joeyscat/qim"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceID       string
	ServiceName     string `default:"wgateway"`
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
	viper.SetConfigType("yaml")

	var config Config

	err := envconfig.Process("qim", &config)
	if err != nil {
		return nil, err
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("read config error: %s", err.Error())
	} else {
		if err := viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	}

	if config.ServiceID == "" {
		localIP := qim.GetLocalIP()
		config.ServiceID = fmt.Sprintf("gate_%s", strings.ReplaceAll(localIP, ".", ""))
	}
	if config.PublicAddress == "" {
		config.PublicAddress = qim.GetLocalIP()
	}

	return &config, nil
}
