package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MinIO     MinIOConfig     `mapstructure:"minio"`
	Redis     RedisConfig     `mapstructure:"redis"`
	FFmpeg    FFmpegConfig    `mapstructure:"ffmpeg"`
	APIServer string          `mapstructure:"api_server"`
	APISecret string          `mapstructure:"api_secret"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	RawBucket string `mapstructure:"raw_bucket"`
	VideoBucket string `mapstructure:"video_bucket"`
	CoverBucket string `mapstructure:"cover_bucket"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type FFmpegConfig struct {
	Binary  string `mapstructure:"binary"`
	Workers int    `mapstructure:"workers"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.SetDefault("server.port", "8081")
	viper.SetDefault("minio.raw_bucket", "raw")
	viper.SetDefault("minio.video_bucket", "video")
	viper.SetDefault("minio.cover_bucket", "cover")
	viper.SetDefault("ffmpeg.binary", "ffmpeg")
	viper.SetDefault("ffmpeg.workers", 2)
	viper.SetDefault("api_secret", "change-me-in-production")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("config file not found, using defaults: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	return &cfg
}
