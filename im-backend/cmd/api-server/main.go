package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"d-im/internal/app"
	"d-im/internal/bootstrap"
	"d-im/internal/config"
	"d-im/pkg/logger"
)

func main() {
	logger.Init()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongoClient.Disconnect(context.Background())

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	db := mongoClient.Database(cfg.MongoDB.Database)
	deps := app.NewDependencies(cfg, db, redisClient, false)

	if err := bootstrap.InitIndexes(cfg); err != nil {
		log.Fatal("Failed to init indexes:", err)
	}

	if err := bootstrap.InitSeed(cfg); err != nil {
		log.Fatal("Failed to init seed:", err)
	}

	r := deps.SetupAPIRouter()

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.APIPort),
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
