package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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

	// 连接 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// 初始化各层组件
	db := mongoClient.Database(cfg.MongoDB.Database)

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
	wsManager := service.NewWSManager(redisClient, sessionService)
	messageService := service.NewMessageService(messageRepo, conversationRepo, conversationService, sessionService, wsManager)
	wsManager.SetMessageService(messageService)

	// 处理器
	messageHandler := handler.NewMessageHandler(messageService, conversationService)
	conversationHandler := handler.NewConversationHandler(conversationService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	wsHandler := handler.NewWSHandler(wsManager, jwtValidator)

	// 初始化 JWT 中间件
	jwtAuthMiddleware := middleware.JWTAuth(cfg, jwtValidator)

	// 启动 WebSocket 管理器
	go wsManager.Run()

	// 设置路由（包含API和WebSocket）
	r := router.SetupAPI(cfg, messageHandler, conversationHandler, sessionHandler, jwtAuthMiddleware)

	// 添加WebSocket路由
	r.GET("/im/ws", wsHandler.HandleWebSocket)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.APIPort),
		Handler: r,
	}

	// 在后台启动服务器
	go func() {
		log.Printf("Unified Server is running on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down unified server...")

	// 创建一个用于超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 WebSocket 管理器
	wsManager.Shutdown()

	// 优雅地关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Unified Server exited")
}
