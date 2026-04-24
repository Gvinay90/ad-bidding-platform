package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	AWS      AWSConfig
	Log      LogConfig
}

type ServerConfig struct {
	CampaignGRPCPort  int `mapstructure:"campaign_grpc_port"`
	BidderGRPCPort    int `mapstructure:"bidder_grpc_port"`
	AnalyticsGRPCPort int `mapstructure:"analytics_grpc_port"`
	GatewayHTTPPort   int `mapstructure:"gateway_http_port"`
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	DNS          string `mapstructure:"dns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type AWSConfig struct {
	Endpoint       string `mapstructure:"endpoint"`
	Region         string `mapstructure:"region"`
	SNSTopic       string `mapstructure:"sns_topic"`
	BidderQueue    string `mapstructure:"bidder_queue"`
	AnalyticsQueue string `mapstructure:"analytics_queue"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
