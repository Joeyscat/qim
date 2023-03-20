package conf

import (
	"encoding/json"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

type Config struct {
	Listen        string `default:":8100"`
	EtcdEndpoints string
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

	return &config, nil
}
