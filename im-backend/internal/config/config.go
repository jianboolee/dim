package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name            string
		DefaultAvatar   string
		FrontendBaseURL string
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
		Secret string
		Expire int
		Issuer string
	}
	Integration struct {
		APIKey string
	}
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("API_SERVER_PORT", 8080)
	viper.SetDefault("WS_SERVER_PORT", 9000)

	viper.SetDefault("MONGODB_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGODB_DATABASE", "x_im")

	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("JWT_SECRET", "")
	viper.SetDefault("JWT_EXPIRE", 3600)
	viper.SetDefault("JWT_ISSUER", "d-im")

	viper.SetDefault("INTEGRATION_API_KEY", "")
	viper.SetDefault("IM_FRONTEND_BASE_URL", "http://localhost:5173")

	viper.SetDefault("APP_NAME", "d-im")
	viper.SetDefault("APP_DEFAULT_AVATAR", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{}

	cfg.Server.APIPort = viper.GetInt("API_SERVER_PORT")
	cfg.Server.WSPort = viper.GetInt("WS_SERVER_PORT")

	cfg.MongoDB.URI = viper.GetString("MONGODB_URI")
	cfg.MongoDB.Database = viper.GetString("MONGODB_DATABASE")

	cfg.Redis.Addr = viper.GetString("REDIS_ADDR")
	cfg.Redis.Password = viper.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = viper.GetInt("REDIS_DB")

	cfg.JWT.Secret = viper.GetString("JWT_SECRET")
	cfg.JWT.Expire = viper.GetInt("JWT_EXPIRE")
	cfg.JWT.Issuer = viper.GetString("JWT_ISSUER")

	cfg.Integration.APIKey = viper.GetString("INTEGRATION_API_KEY")
	cfg.App.FrontendBaseURL = viper.GetString("IM_FRONTEND_BASE_URL")

	cfg.App.Name = viper.GetString("APP_NAME")
	cfg.App.DefaultAvatar = viper.GetString("APP_DEFAULT_AVATAR")
	if cfg.App.DefaultAvatar == "" {
		cfg.App.DefaultAvatar = viper.GetString("DEFAULT_AVATAR")
	}

	return cfg, nil
}
