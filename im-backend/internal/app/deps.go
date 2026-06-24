package app

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"d-im/internal/config"
	"d-im/internal/handler"
	"d-im/internal/middleware"
	"d-im/internal/repository"
	"d-im/internal/router"
	"d-im/internal/service"
	jwtpkg "d-im/pkg/jwt"
)

type Dependencies struct {
	Config                 *config.Config
	JWTService             *jwtpkg.Service
	JWTAuthMiddleware      gin.HandlerFunc
	IntegrationMiddleware  gin.HandlerFunc
	MessageHandler         *handler.MessageHandler
	ConversationHandler    *handler.ConversationHandler
	SessionHandler         *handler.SessionHandler
	UserHandler            *handler.UserHandler
	IntegrationHandler     *handler.IntegrationHandler
	WSHandler              *handler.WSHandler
	WSManager              *service.WSManager
}

func NewDependencies(cfg *config.Config, db *mongo.Database, redisClient *redis.Client, withWS bool) *Dependencies {
	jwtService, err := jwtpkg.InitService(cfg.JWT.Secret, cfg.JWT.Expire, cfg.JWT.Issuer)
	if err != nil {
		log.Fatal("Failed to initialize JWT service:", err)
	}

	messageRepo := repository.NewMessageRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	userRepo := repository.NewUserRepository(db)

	sessionService := service.NewSessionService(sessionRepo)
	conversationService := service.NewConversationService(conversationRepo)
	userService := service.NewUserService(userRepo)
	integrationService := service.NewIntegrationService(
		userService,
		conversationService,
		jwtService,
		cfg.App.FrontendBaseURL,
	)

	var wsManager *service.WSManager
	var messageService *service.MessageService

	if withWS {
		wsManager = service.NewWSManager(redisClient, sessionService)
		messageService = service.NewMessageService(messageRepo, conversationRepo, conversationService, sessionService, wsManager)
		wsManager.SetMessageService(messageService)
	} else {
		messageService = service.NewMessageService(messageRepo, conversationRepo, conversationService, sessionService, nil)
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
		IntegrationHandler:    handler.NewIntegrationHandler(integrationService),
		WSManager:             wsManager,
	}

	if withWS && wsManager != nil {
		deps.WSHandler = handler.NewWSHandler(wsManager, jwtService)
	}

	return deps
}

func (d *Dependencies) SetupAPIRouter() *gin.Engine {
	return router.SetupAPI(
		d.Config,
		d.MessageHandler,
		d.ConversationHandler,
		d.SessionHandler,
		d.UserHandler,
		d.IntegrationHandler,
		d.JWTAuthMiddleware,
		d.IntegrationMiddleware,
	)
}

func (d *Dependencies) SetupWSRouter() *gin.Engine {
	return router.SetupWS(d.Config, d.WSHandler)
}
