package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"d-im/internal/bootstrap"
	"d-im/internal/config"
	"d-im/internal/handler"
	"d-im/internal/middleware"
	"d-im/internal/repository"
	"d-im/internal/router"
	"d-im/internal/service"
	jwtpkg "d-im/pkg/jwt"
	"d-im/pkg/logger"
)

func main() {
	logger.Init()

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 连接 MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongoClient.Disconnect(context.Background())
	db := mongoClient.Database(cfg.MongoDB.Database)

	// 连接 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// 初始化 JWT
	jwtValidator, err := jwtpkg.InitJwtValidator(cfg.JWT.PublicKeyPath)
	if err != nil {
		log.Fatal("Failed to initialize JWT validator:", err)
	}

	// 仓储层
	messageRepo := repository.NewMessageRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// 服务层
	sessionService := service.NewSessionService(sessionRepo)
	conversationService := service.NewConversationService(conversationRepo)
	messageService := service.NewMessageService(messageRepo, conversationRepo, conversationService, sessionService, nil) // API服务不需要WSManager

	// 处理器
	messageHandler := handler.NewMessageHandler(messageService, conversationService)
	conversationHandler := handler.NewConversationHandler(conversationService)
	sessionHandler := handler.NewSessionHandler(sessionService)

	// 初始化 JWT 中间件
	jwtAuthMiddleware := middleware.JWTAuth(cfg, jwtValidator)

	// 设置路由
	r := router.SetupAPI(cfg, messageHandler, conversationHandler, sessionHandler, jwtAuthMiddleware)

	// 初始化索引
	if err := bootstrap.InitIndexes(cfg); err != nil {
		log.Fatal("Failed to init indexes:", err)
	}

	// 创建默认用户
	if err := bootstrap.InitSeed(cfg); err != nil {
		log.Fatal("Failed to init seed:", err)
	}

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.APIPort),
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
