package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Chat      ChatConfig      `mapstructure:"chat"`
	Ilo       IloConfig       `mapstructure:"ilo"`
	LLM       LLMConfig       `mapstructure:"llm"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Tracing   TracingConfig   `mapstructure:"tracing"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type AuthConfig struct {
	ServiceAddr     string        `mapstructure:"service_addr"`
	JWTSecret       string        `mapstructure:"jwt_secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

type ChatConfig struct {
	ServiceAddr string `mapstructure:"service_addr"`
}

type IloConfig struct {
	ServiceAddr string `mapstructure:"service_addr"`
}

type LLMConfig struct {
	ServiceAddr string `mapstructure:"service_addr"`
}

type RateLimitConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	RequestsPerMinute int    `mapstructure:"requests_per_minute"`
	RedisAddr         string `mapstructure:"redis_addr"`
}

type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	ServiceName string `mapstructure:"service_name"`
	Endpoint    string `mapstructure:"endpoint"`
	Insecure    bool   `mapstructure:"insecure"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
