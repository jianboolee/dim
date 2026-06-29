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
	if err := InitConversationMemberIndexes(context.Background(), db.Database(cfg.MongoDB.Database)); err != nil {
		return err
	}
	if err := InitGroupIndexes(context.Background(), db.Database(cfg.MongoDB.Database)); err != nil {
		return err
	}
	if err := InitGroupMemberIndexes(context.Background(), db.Database(cfg.MongoDB.Database)); err != nil {
		return err
	}
	if err := InitSessionIndexes(context.Background(), db.Database(cfg.MongoDB.Database)); err != nil {
		return err
	}
	if err := InitAuthSessionIndexes(context.Background(), db.Database(cfg.MongoDB.Database)); err != nil {
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
			Keys: bson.D{{Key: "conversation_id", Value: 1}, {Key: "seq", Value: -1}},
			Options: options.Index().
				SetName("idx_conversation_seq_desc"),
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
				SetPartialFilterExpression(bson.M{"client_message_id": bson.M{"$exists": true, "$gt": ""}}).
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
			Keys: bson.D{{Key: "hash_id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"hash_id": bson.M{"$exists": true, "$gt": ""}}).
				SetName("unique_hash_id"),
		},
		{
			Keys: bson.D{{Key: "group_id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetSparse(true).
				SetName("unique_group_id"),
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

func InitConversationMemberIndexes(ctx context.Context, db *mongo.Database) error {
	memberCollection := db.Collection(models.CollectionConversationMember)

	return EnsureIndexes(ctx, memberCollection, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "conversation_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("unique_conversation_member"),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}, {Key: "sort_at", Value: -1}},
			Options: options.Index().
				SetName("idx_conversation_member_user_status_sort"),
		},
		{
			Keys: bson.D{{Key: "conversation_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().
				SetName("idx_conversation_member_conversation_status"),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "unread_count", Value: 1}},
			Options: options.Index().
				SetName("idx_conversation_member_user_unread"),
		},
	})
}

func InitGroupIndexes(ctx context.Context, db *mongo.Database) error {
	groupCollection := db.Collection(models.CollectionGroup)

	return EnsureIndexes(ctx, groupCollection, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "conversation_id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("unique_group_conversation_id"),
		},
		{
			Keys: bson.D{{Key: "owner_id", Value: 1}},
			Options: options.Index().
				SetName("idx_group_owner_id"),
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
			Options: options.Index().
				SetName("idx_group_status"),
		},
	})
}

func InitGroupMemberIndexes(ctx context.Context, db *mongo.Database) error {
	groupMemberCollection := db.Collection(models.CollectionGroupMember)

	return EnsureIndexes(ctx, groupMemberCollection, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("unique_group_member"),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().
				SetName("idx_group_member_user_status"),
		},
		{
			Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "role", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().
				SetName("idx_group_member_role_status"),
		},
	})
}

// InitIndexes 初始化索引, 给 last_seen 字段创建 TTL 索引, 设置自动过期时间
func InitSessionIndexes(ctx context.Context, db *mongo.Database) error {
	sessionCollection := db.Collection(models.CollectionSession)
	const (
		sessionTTLIndexName     = "last_seen_1"
		sessionTTLExpireSeconds = int32(60 * 5)
	)

	if err := dropConflictingTTLIndex(ctx, sessionCollection, sessionTTLIndexName, sessionTTLExpireSeconds); err != nil {
		return fmt.Errorf("prepare session ttl index failed: %w", err)
	}

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "last_seen", Value: 1}},
		Options: options.Index().
			SetName(sessionTTLIndexName).
			SetExpireAfterSeconds(sessionTTLExpireSeconds), // 5 分钟后过期
	}

	name, err := sessionCollection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("init session indexes failed: %w", err)
	}

	log.Printf("Indexes ensured for [%s]: [%s]\n", sessionCollection.Name(), name)
	return nil
}

func dropConflictingTTLIndex(ctx context.Context, collection *mongo.Collection, indexName string, expireSeconds int32) error {
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			return err
		}
		if index["name"] != indexName {
			continue
		}

		if indexTTLSeconds(index["expireAfterSeconds"]) == expireSeconds {
			return nil
		}

		if _, err := collection.Indexes().DropOne(ctx, indexName); err != nil {
			return err
		}
		log.Printf("Dropped conflicting TTL index [%s.%s]\n", collection.Name(), indexName)
		return nil
	}

	return cursor.Err()
}

func indexTTLSeconds(value any) int32 {
	switch v := value.(type) {
	case int32:
		return v
	case int64:
		return int32(v)
	case int:
		return int32(v)
	case float64:
		return int32(v)
	default:
		return -1
	}
}

func InitAuthSessionIndexes(ctx context.Context, db *mongo.Database) error {
	authSessionCollection := db.Collection(models.CollectionAuthSession)

	return EnsureIndexes(ctx, authSessionCollection, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("unique_auth_session_id"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("idx_auth_session_user_updated"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "last_active_at", Value: -1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_auth_session_user_active"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "device_id", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("idx_auth_session_user_device"),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("ttl_auth_session_expiry"),
		},
	})
}
