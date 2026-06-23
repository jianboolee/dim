package bootstrap

import (
	"context"
	"d-im/internal/config"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestMongoConnection 尝试连接并 ping MongoDB，成功则返回 mongo.Client 实例
func TestMongoConnection(cfg *config.Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.MongoDB.URI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("MongoDB ping failed: %w", err)
	}

	return client, nil
}
