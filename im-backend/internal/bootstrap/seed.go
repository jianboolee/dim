package bootstrap

import (
	"context"
	"d-im/internal/config"
	"d-im/internal/models"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 数据库迁移
func InitSeed(cfg *config.Config) error {

	// 连接数据库
	db, err := TestMongoConnection(cfg)
	if err != nil {
		return err
	}
	defer db.Disconnect(context.Background())

	// 创建默认用户
	if err := seedUser(context.Background(), db, cfg); err != nil {
		return err
	}

	// 创建默认会话
	if err := seedConversation(context.Background(), db, cfg); err != nil {
		return err
	}

	return nil
}

// 创建默认用户
func seedUser(ctx context.Context, db *mongo.Client, cfg *config.Config) error {
	userCollection := db.Database(cfg.MongoDB.Database).Collection(models.CollectionUser)

	// 创建默认用户
	user := models.User{
		ID:        models.SystemUserID,
		Nickname:  "System",
		Avatar:    cfg.App.DefaultAvatar,
		Bio:       "This is a system user",
		CreatedAt: time.Now(),
	}
	// 尝试插入，如果已存在就忽略
	_, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		if writeErr, ok := err.(mongo.WriteException); ok {
			for _, e := range writeErr.WriteErrors {
				if e.Code == 11000 {
					// 说明用户已存在，跳过
					log.Println("Default user already exists, skipping...")
					return nil
				}
			}
		}
		return fmt.Errorf("failed to insert default user: %w", err)
	}

	log.Println("Default user created successfully.")

	return nil
}

func seedConversation(ctx context.Context, db *mongo.Client, cfg *config.Config) error {
	conversationCollection := db.Database(cfg.MongoDB.Database).Collection(models.CollectionConversation)

	exist, err := conversationCollection.CountDocuments(ctx, bson.M{"type": models.ConversationTypeSystem})
	if err != nil {
		return fmt.Errorf("failed to check if conversation exists: %w", err)
	}
	if exist > 0 {
		log.Println("Default conversation already exists, skipping...")
		return nil
	}

	conversation := models.Conversation{
		ID:           primitive.NewObjectID(),
		Type:         models.ConversationTypeSystem,
		Participants: []string{models.SystemUserID},
		UnreadCounts: map[string]int64{models.SystemUserID: 0},
		LastMessage:  nil,
		CreatedAt:    time.Now(),
	}

	inserted, err := conversationCollection.InsertOne(ctx, conversation)
	if err != nil {
		return fmt.Errorf("failed to insert conversation: %w", err)
	}

	log.Printf("Default conversation created successfully. ID: %s", inserted)

	return nil
}
