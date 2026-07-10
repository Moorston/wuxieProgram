package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Mongo    MongoConfig    `mapstructure:"mongo"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	WX       WXConfig       `mapstructure:"wx"`
	CORS     CORSConfig     `mapstructure:"cors"`
	MediaURL    string         `mapstructure:"media_url"`
	MediaSecret string         `mapstructure:"media_secret"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MongoConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret  string `mapstructure:"secret"`
	Expires int    `mapstructure:"expires"`
}

type WXConfig struct {
	AppID      string `mapstructure:"app_id"`
	Secret     string `mapstructure:"secret"`
	TemplateID string `mapstructure:"template_id"`
	RemindHour int    `mapstructure:"remind_hour"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("jwt.expires", 72)
	viper.SetDefault("cors.allowed_origins", []string{
		"http://localhost:8080",
		"http://localhost:3000",
		"http://localhost",
	})

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("config file not found, using defaults: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	// 验证关键配置项
	if cfg.Mongo.URI == "" {
		log.Fatal("FATAL: mongo.uri is required")
	}
	if cfg.JWT.Secret == "" {
		log.Fatal("FATAL: jwt.secret is required")
	}

	return &cfg
}
