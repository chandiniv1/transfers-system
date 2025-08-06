package util

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBDriver       string `mapstructure:"DB_DRIVER"`
	DBSource       string `mapstructure:"DB_SOURCE"`
	MigrationURL   string `mapstructure:"MIGRATION_URL"`
	Environment    string `mapstructure:"ENVIRONMENT"`
	HTTPServerAddr string `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddr string `mapstructure:"GRPC_SERVER_ADDRESS"`
}

// LoadConfig reads configuration from file or environment variable
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
