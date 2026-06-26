package app

import (
	"context"
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"d-im/internal/config"
	"d-im/internal/handler"
	"d-im/internal/middleware"
	"d-im/internal/repository"
	"d-im/internal/router"
	"d-im/internal/service"
	"d-im/internal/upload"
	jwtpkg "d-im/pkg/jwt"
)

type Dependencies struct {
	Config                *config.Config
	JWTService            *jwtpkg.Service
	JWTAuthMiddleware     gin.HandlerFunc
	IntegrationMiddleware gin.HandlerFunc
	MessageHandler        *handler.MessageHandler
	ConversationHandler   *handler.ConversationHandler
	SessionHandler        *handler.SessionHandler
	UserHandler           *handler.UserHandler
	AuthHandler           *handler.AuthHandler
	IntegrationHandler    *handler.IntegrationHandler
	UploadHandler         *upload.Handler
	WSHandler             *handler.WSHandler
	WSManager             *service.WSManager
}

func NewDependencies(cfg *config.Config, db *mongo.Database, redisClient *redis.Client, withWS bool) *Dependencies {
	if redisClient != nil {
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			log.Printf("WARNING: Redis ping failed: %v (split API/WS real-time push requires Redis)", err)
		}
	}

	jwtService, err := jwtpkg.InitService(
		cfg.JWT.Secret,
		cfg.JWT.Expire,
		cfg.JWT.RefreshExpire,
		cfg.JWT.MaxSession,
		cfg.JWT.Issuer,
	)
	if err != nil {
		log.Fatal("Failed to initialize JWT service:", err)
	}

	messageRepo := repository.NewMessageRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	userRepo := repository.NewUserRepository(db)
	authSessionRepo := repository.NewAuthSessionRepository(db)

	sessionService := service.NewSessionService(sessionRepo)
	conversationService := service.NewConversationService(conversationRepo, userRepo)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(jwtService, authSessionRepo, service.AuthCookieConfig{
		Name:     cfg.JWT.RefreshCookieName,
		Domain:   cfg.JWT.RefreshCookieDomain,
		Secure:   cfg.JWT.RefreshCookieSecure,
		SameSite: service.ParseSameSite(cfg.JWT.RefreshCookieSameSite),
	})
	integrationService := service.NewIntegrationService(
		userService,
		conversationService,
		authService,
		cfg.App.FrontendBaseURL,
	)
	uploadHandler := newUploadHandler(cfg)

	var wsManager *service.WSManager
	var messageService *service.MessageService

	if withWS {
		wsManager = service.NewWSManager(redisClient, sessionService, service.WSManagerOptions{
			PingPeriod: cfg.WebSocket.HeartbeatInterval,
			PongWait:   cfg.WebSocket.PongTimeout,
			WriteWait:  cfg.WebSocket.WriteTimeout,
		})
		messageService = service.NewMessageService(messageRepo, conversationRepo, conversationService, sessionService, wsManager, redisClient)
		wsManager.SetMessageService(messageService)
	} else {
		messageService = service.NewMessageService(messageRepo, conversationRepo, conversationService, sessionService, nil, redisClient)
	}

	deps := &Dependencies{
		Config:                cfg,
		JWTService:            jwtService,
		JWTAuthMiddleware:     middleware.JWTAuth(jwtService),
		IntegrationMiddleware: middleware.IntegrationAPIKey(cfg),
		MessageHandler:        handler.NewMessageHandler(messageService, conversationService),
		ConversationHandler:   handler.NewConversationHandler(conversationService),
		SessionHandler:        handler.NewSessionHandler(sessionService),
		UserHandler:           handler.NewUserHandler(userService),
		AuthHandler:           handler.NewAuthHandler(authService),
		IntegrationHandler:    handler.NewIntegrationHandler(integrationService),
		UploadHandler:         uploadHandler,
		WSManager:             wsManager,
	}

	if withWS && wsManager != nil {
		deps.WSHandler = handler.NewWSHandler(wsManager, jwtService)
	}

	return deps
}

func newUploadHandler(cfg *config.Config) *upload.Handler {
	storageCfg := &upload.StorageConfig{
		Endpoint:        cfg.Storage.OSSEndpoint,
		AccessKeyID:     cfg.Storage.OSSAccessKeyID,
		AccessKeySecret: cfg.Storage.OSSAccessKeySecret,
		BucketName:      cfg.Storage.OSSBucketName,
		CustomDomain:    cfg.Storage.OSSCustomDomain,
		Directory:       cfg.Storage.OSSDirectory,
	}

	ossClient, err := upload.NewOSSClient(storageCfg, slog.Default())
	if err != nil {
		log.Printf("WARNING: upload storage disabled: %v", err)
	}

	return upload.NewHandler(upload.NewService(ossClient))
}

func (d *Dependencies) SetupAPIRouter() *gin.Engine {
	return router.SetupAPI(
		d.Config,
		d.MessageHandler,
		d.ConversationHandler,
		d.SessionHandler,
		d.UserHandler,
		d.AuthHandler,
		d.IntegrationHandler,
		d.UploadHandler,
		d.JWTAuthMiddleware,
		d.IntegrationMiddleware,
	)
}

func (d *Dependencies) SetupWSRouter() *gin.Engine {
	return router.SetupWS(d.Config, d.WSHandler)
}
