package bootstrap

import (
	"context"
	"d-im/internal/config"
	"d-im/internal/models"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 数据库迁移
func InitIndexes(cfg *config.Config) error {
	// 连接数据库
	db, err := TestMongoConnection(cfg)
	if err != nil {
		return err
	}
	defer db.Disconnect(context.Background())

	// 初始化索引
	if err := InitUserIndexes(context.Background(), db.Database(cfg.MongoDB.Database)); err != nil {
		return err
	}
	if err := InitMessageIndexes(context.Background(), db.Database(cfg.MongoDB.Database)); err != nil {
		return err
	}
	if err := InitConversationIndexes(context.Background(), db.Database(cfg.MongoDB.Database)); err != nil {
		return err
	}
	if err := InitSessionIndexes(context.Background(), db.Database(cfg.MongoDB.Database)); err != nil {
		return err
	}

	return nil
}

// EnsureIndexes 为给定集合创建索引（可传多个）
func EnsureIndexes(ctx context.Context, collection *mongo.Collection, indexes []mongo.IndexModel) error {
	if len(indexes) == 0 {
		return nil
	}

	names, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("create indexes failed for %s: %w", collection.Name(), err)
	}

	log.Printf("Indexes ensured for [%s]: %v\n", collection.Name(), names)
	return nil
}

// 初始化用户索引
func InitUserIndexes(ctx context.Context, db *mongo.Database) error {
	userCollection := db.Collection(models.CollectionUser)

	return EnsureIndexes(ctx, userCollection, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}}, // 外部系统 ID
			Options: options.Index().
				SetUnique(true).
				SetName("unique_user_id"),
		},
		{
			Keys: bson.D{{Key: "nickname", Value: 1}},
			Options: options.Index().
				SetName("idx_user_nickname"),
		},
	})
}

// 初始化消息索引
func InitMessageIndexes(ctx context.Context, db *mongo.Database) error {
	messageCollection := db.Collection(models.CollectionMessage)
	return EnsureIndexes(ctx, messageCollection, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "conversation_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().
				SetName("idx_conversation_created_at"),
		},
		{
			Keys:    bson.D{{Key: "sender_id", Value: 1}},
			Options: options.Index().SetName("idx_sender"),
		},
		{
			Keys: bson.D{
				{Key: "conversation_id", Value: 1},
				{Key: "sender_id", Value: 1},
				{Key: "client_message_id", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"client_message_id": bson.M{"$exists": true, "$ne": ""}}).
				SetName("unique_message_client_id"),
		},
	})
}

// 初始化会话索引
func InitConversationIndexes(ctx context.Context, db *mongo.Database) error {
	conversationCollection := db.Collection(models.CollectionConversation)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "participants", Value: 1}},
			Options: options.Index().
				SetName("idx_participants"),
		},
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "participants", Value: 1},
			},
			Options: options.Index().
				SetName("idx_type_participants"),
		},
		{
			Keys:    bson.D{{Key: "hash_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("unique_hash_id"),
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: -1}},
			Options: options.Index().
				SetName("idx_updated_at_desc"),
		},
		{
			Keys: bson.D{
				{Key: "participants", Value: 1},
				{Key: "updated_at", Value: -1},
				{Key: "_id", Value: -1},
			},
			Options: options.Index().
				SetName("idx_participants_updated_id_desc"),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().
				SetName("idx_created_at_desc"),
		},
	}

	if err := EnsureIndexes(ctx, conversationCollection, indexes); err != nil {
		return fmt.Errorf("init conversation indexes failed: %w", err)
	}
	return nil
}

// InitIndexes 初始化索引, 给 last_seen 字段创建 TTL 索引, 设置自动过期时间
func InitSessionIndexes(ctx context.Context, db *mongo.Database) error {
	sessionCollection := db.Collection(models.CollectionSession)

	indexModel := mongo.IndexModel{
		Keys: bson.M{"last_seen": 1},
		Options: options.Index().
			SetExpireAfterSeconds(60 * 60 * 24), // 24 小时后过期（你可改成 300 秒等）
	}

	fmt.Printf("Indexes ensured for %s\n", sessionCollection.Name())

	_, err := sessionCollection.Indexes().CreateOne(ctx, indexModel)

	if err != nil {
		return fmt.Errorf("init session indexes failed: %w", err)
	}

	return nil
}
