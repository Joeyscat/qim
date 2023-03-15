package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var (
	L *zap.Logger
)

type Settings struct {
	Env         string
	Filename    string
	Format      string
	Level       string
	RollingDays uint
}

func Init(settings Settings) error {
	if settings.Env == "" {
		settings.Env = "dev"
	}

	if settings.Filename == "" {
		settings.Filename = "qim.log"
	}
	if settings.Format == "" {
		settings.Format = "console"
	}

	if settings.Level == "" {
		settings.Level = "debug"
	}

	if settings.RollingDays == 0 {
		settings.RollingDays = 7
	}

	level := new(zap.AtomicLevel)
	err := level.UnmarshalText([]byte(settings.Level))
	if err != nil {
		return err
	}

	if settings.Env == "dev" {
		L, err = zap.NewDevelopment()
	} else if settings.Env == "prod" {
		L, err = zap.NewProduction()
	} else {
		return fmt.Errorf("unsupported env: %s", settings.Env)
	}
	return err
}
