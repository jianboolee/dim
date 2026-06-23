package repository

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"d-im/internal/models"
)

type SessionRepository struct {
	db                *mongo.Database
	sessionCollection *mongo.Collection
}

func NewSessionRepository(db *mongo.Database) *SessionRepository {
	return &SessionRepository{db: db, sessionCollection: db.Collection(models.CollectionSession)}
}

// UpsertSession 更新或创建用户会话
func (r *SessionRepository) UpsertSession(ctx context.Context, session *models.Session) error {
	filter := bson.M{"user_id": session.UserID}
	update := bson.M{
		"$set": bson.M{
			"is_online": session.IsOnline,
			"last_seen": session.LastSeen,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := r.sessionCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("Error upserting session: %v", err)
		return err
	}

	return nil
}

// GetSession 获取用户会话信息
func (r *SessionRepository) GetSession(ctx context.Context, userID string) (*models.Session, error) {
	var session models.Session
	err := r.sessionCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&session)
	if err == mongo.ErrNoDocuments {
		// 如果没有找到会话记录，返回一个新的离线会话
		return &models.Session{
			UserID:   userID,
			IsOnline: false,
			LastSeen: time.Time{},
		}, nil
	}
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return nil, err
	}

	return &session, nil
}

// GetMultiSession 批量获取用户会话信息
func (r *SessionRepository) GetMultiSession(ctx context.Context, userIDs []string) (map[string]*models.Session, error) {
	cursor, err := r.sessionCollection.Find(ctx, bson.M{
		"user_id": bson.M{"$in": userIDs},
	})
	if err != nil {
		log.Printf("Error finding sessions: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	sessions := make(map[string]*models.Session)
	for cursor.Next(ctx) {
		var session models.Session
		if err := cursor.Decode(&session); err != nil {
			log.Printf("Error decoding session: %v", err)
			continue
		}
		sessions[session.UserID] = &session
	}

	// 为没有会话记录的用户创建离线会话
	for _, userID := range userIDs {
		if _, exists := sessions[userID]; !exists {
			sessions[userID] = &models.Session{
				UserID:   userID,
				IsOnline: false,
				LastSeen: time.Time{},
			}
		}
	}

	return sessions, nil
}

// UpdateLastSeen 更新用户最后在线时间
func (r *SessionRepository) UpdateLastSeen(ctx context.Context, userID string) error {
	update := bson.M{
		"$set": bson.M{
			"last_seen": time.Now(),
		},
	}

	_, err := r.sessionCollection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
	if err != nil {
		log.Printf("Error updating last seen: %v", err)
		return err
	}

	return nil
}

// 获取在线用户总数
func (r *SessionRepository) GetOnlineUserCount(ctx context.Context) (int64, error) {
	count, err := r.sessionCollection.CountDocuments(ctx, bson.M{"is_online": true})
	if err != nil {
		log.Printf("Error getting online user count: %v", err)
		return 0, err
	}
	return count, nil
}
