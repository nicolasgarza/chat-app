package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Logger    LoggerConfig
	RateLimit RateLimitConfig
}

type LoggerConfig struct {
	OutputPaths      []string
	ErrorOutputPaths []string
	Level            string
}

type RateLimitConfig struct {
	Rate  time.Duration
	Burst int
}

var AppConfig *Config

func LoadConfig() error {
	v := viper.New()

	v.SetDefault("logger.outputPaths", []string{"stdout"})
	v.SetDefault("logger.errorOutputPaths", []string{"stderr"})
	v.SetDefault("logger.level", "info")

	v.SetDefault("ratelimit.rate", "1s")
	v.SetDefault("logger.level", 100)

	v.AutomaticEnv()

	AppConfig = &Config{}
	if err := v.Unmarshal(AppConfig); err != nil {
		return err
	}

	// Parse the rate string to time.Duration
	rate, err := time.ParseDuration(v.GetString("ratelimit.rate"))
	if err != nil {
		return err
	}
	AppConfig.RateLimit.Rate = rate

	return nil
}
