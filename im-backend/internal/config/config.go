package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name          string
		DefaultAvatar string
	}
	Server struct {
		APIPort int
		WSPort  int
	}
	MongoDB struct {
		URI      string
		Database string
	}
	Redis struct {
		Addr     string
		Password string
		DB       int
	}
	JWT struct {
		PublicKeyPath string
	}
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// 设置默认值
	viper.SetDefault("API_SERVER_PORT", 8080)
	viper.SetDefault("WS_SERVER_PORT", 9000)

	viper.SetDefault("MONGODB_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGODB_DATABASE", "x_im")

	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("JWT_PUBLIC_KEY_PATH", "./config/keys/public.pem")

	viper.SetDefault("APP_NAME", "XIM")
	viper.SetDefault("APP_DEFAULT_AVATAR", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{}

	// 从环境变量读取配置
	cfg.Server.APIPort = viper.GetInt("API_SERVER_PORT")
	cfg.Server.WSPort = viper.GetInt("WS_SERVER_PORT")

	cfg.MongoDB.URI = viper.GetString("MONGODB_URI")
	cfg.MongoDB.Database = viper.GetString("MONGODB_DATABASE")

	cfg.Redis.Addr = viper.GetString("REDIS_ADDR")
	cfg.Redis.Password = viper.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = viper.GetInt("REDIS_DB")

	cfg.JWT.PublicKeyPath = viper.GetString("JWT_PUBLIC_KEY_PATH")

	cfg.App.Name = viper.GetString("APP_NAME")
	cfg.App.DefaultAvatar = viper.GetString("APP_DEFAULT_AVATAR")

	return cfg, nil
}
