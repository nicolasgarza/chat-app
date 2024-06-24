package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Logger LoggerConfig
}

type LoggerConfig struct {
	OutputPaths      []string
	ErrorOutputPaths []string
	Level            string
}

var AppConfig *Config

func LoadConfig() error {
	v := viper.New()

	v.SetDefault("logger.outputPaths", []string{"stdout"})
	v.SetDefault("logger.errorOutputPaths", []string{"stderr"})
	v.SetDefault("logger.level", "info")

	v.AutomaticEnv()

	AppConfig = &Config{}
	if err := v.Unmarshal(AppConfig); err != nil {
		return err
	}

	return nil
}
