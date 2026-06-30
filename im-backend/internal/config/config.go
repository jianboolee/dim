package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name            string
		DefaultAvatar   string
		FrontendBaseURL string
		PublicBaseURL   string
	}
	Server struct {
		APIPort int
		WSPort  int
	}
	WebSocket struct {
		HeartbeatInterval time.Duration
		PongTimeout       time.Duration
		WriteTimeout      time.Duration
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
		Secret                string
		Expire                time.Duration
		RefreshExpire         time.Duration
		MaxSession            time.Duration
		Issuer                string
		RefreshCookieName     string
		RefreshCookieDomain   string
		RefreshCookieSecure   bool
		RefreshCookieSameSite string
		MaxActiveSessions     int64
	}
	Integration struct {
		APIKey string
	}
	Storage struct {
		OSSEndpoint        string
		OSSAccessKeyID     string
		OSSAccessKeySecret string
		OSSBucketName      string
		OSSCustomDomain    string
		OSSDirectory       string
	}
	GroupAvatar struct {
		Directory    string
		Size         int
		KeepVersions int
	}
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("API_SERVER_PORT", 8080)
	viper.SetDefault("WS_SERVER_PORT", 9000)
	viper.SetDefault("WS_HEARTBEAT_INTERVAL", "60s")
	viper.SetDefault("WS_PONG_TIMEOUT", "75s")
	viper.SetDefault("WS_WRITE_TIMEOUT", "10s")

	viper.SetDefault("MONGODB_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGODB_DATABASE", "x_im")

	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("JWT_SECRET", "")
	viper.SetDefault("JWT_EXPIRE", "1h")
	viper.SetDefault("JWT_REFRESH_EXPIRE", "168h")
	viper.SetDefault("JWT_MAX_SESSION", "168h")
	viper.SetDefault("JWT_ISSUER", "d-im")
	viper.SetDefault("JWT_REFRESH_COOKIE_NAME", "d_im_refresh_token")
	viper.SetDefault("JWT_REFRESH_COOKIE_DOMAIN", "")
	viper.SetDefault("JWT_REFRESH_COOKIE_SECURE", false)
	viper.SetDefault("JWT_REFRESH_COOKIE_SAMESITE", "Lax")
	viper.SetDefault("JWT_MAX_ACTIVE_SESSIONS", 10)

	viper.SetDefault("INTEGRATION_API_KEY", "")
	viper.SetDefault("IM_FRONTEND_BASE_URL", "http://localhost:5173")
	viper.SetDefault("IM_PUBLIC_BASE_URL", "http://localhost:8080")

	viper.SetDefault("APP_NAME", "d-im")
	viper.SetDefault("APP_DEFAULT_AVATAR", "")

	viper.SetDefault("OSS_ENDPOINT", "")
	viper.SetDefault("OSS_ACCESS_KEY_ID", "")
	viper.SetDefault("OSS_ACCESS_KEY_SECRET", "")
	viper.SetDefault("OSS_BUCKET_NAME", "")
	viper.SetDefault("OSS_CUSTOM_DOMAIN", "")
	viper.SetDefault("OSS_DIRECTORY", "uploads")
	viper.SetDefault("IM_GROUP_AVATAR_DIR", "./storage/group-avatars")
	viper.SetDefault("IM_GROUP_AVATAR_SIZE", 240)
	viper.SetDefault("IM_GROUP_AVATAR_KEEP_VERSIONS", 3)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !os.IsNotExist(err) {
			return nil, err
		}
	}

	cfg := &Config{}

	cfg.Server.APIPort = viper.GetInt("API_SERVER_PORT")
	cfg.Server.WSPort = viper.GetInt("WS_SERVER_PORT")
	cfg.WebSocket.HeartbeatInterval = viper.GetDuration("WS_HEARTBEAT_INTERVAL")
	cfg.WebSocket.PongTimeout = viper.GetDuration("WS_PONG_TIMEOUT")
	cfg.WebSocket.WriteTimeout = viper.GetDuration("WS_WRITE_TIMEOUT")

	cfg.MongoDB.URI = viper.GetString("MONGODB_URI")
	cfg.MongoDB.Database = viper.GetString("MONGODB_DATABASE")

	cfg.Redis.Addr = viper.GetString("REDIS_ADDR")
	cfg.Redis.Password = viper.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = viper.GetInt("REDIS_DB")

	cfg.JWT.Secret = viper.GetString("JWT_SECRET")
	cfg.JWT.Expire = viper.GetDuration("JWT_EXPIRE")
	cfg.JWT.RefreshExpire = viper.GetDuration("JWT_REFRESH_EXPIRE")
	cfg.JWT.MaxSession = viper.GetDuration("JWT_MAX_SESSION")
	cfg.JWT.Issuer = viper.GetString("JWT_ISSUER")
	cfg.JWT.RefreshCookieName = viper.GetString("JWT_REFRESH_COOKIE_NAME")
	cfg.JWT.RefreshCookieDomain = viper.GetString("JWT_REFRESH_COOKIE_DOMAIN")
	cfg.JWT.RefreshCookieSecure = viper.GetBool("JWT_REFRESH_COOKIE_SECURE")
	cfg.JWT.RefreshCookieSameSite = viper.GetString("JWT_REFRESH_COOKIE_SAMESITE")
	cfg.JWT.MaxActiveSessions = viper.GetInt64("JWT_MAX_ACTIVE_SESSIONS")

	cfg.Integration.APIKey = viper.GetString("INTEGRATION_API_KEY")
	cfg.App.FrontendBaseURL = viper.GetString("IM_FRONTEND_BASE_URL")
	cfg.App.PublicBaseURL = viper.GetString("IM_PUBLIC_BASE_URL")

	cfg.App.Name = viper.GetString("APP_NAME")
	cfg.App.DefaultAvatar = viper.GetString("APP_DEFAULT_AVATAR")
	if cfg.App.DefaultAvatar == "" {
		cfg.App.DefaultAvatar = viper.GetString("DEFAULT_AVATAR")
	}

	cfg.Storage.OSSEndpoint = viper.GetString("OSS_ENDPOINT")
	cfg.Storage.OSSAccessKeyID = viper.GetString("OSS_ACCESS_KEY_ID")
	cfg.Storage.OSSAccessKeySecret = viper.GetString("OSS_ACCESS_KEY_SECRET")
	cfg.Storage.OSSBucketName = viper.GetString("OSS_BUCKET_NAME")
	cfg.Storage.OSSCustomDomain = viper.GetString("OSS_CUSTOM_DOMAIN")
	cfg.Storage.OSSDirectory = viper.GetString("OSS_DIRECTORY")
	cfg.GroupAvatar.Directory = viper.GetString("IM_GROUP_AVATAR_DIR")
	cfg.GroupAvatar.Size = viper.GetInt("IM_GROUP_AVATAR_SIZE")
	cfg.GroupAvatar.KeepVersions = viper.GetInt("IM_GROUP_AVATAR_KEEP_VERSIONS")

	return cfg, nil
}
