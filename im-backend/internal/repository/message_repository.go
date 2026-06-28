package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"d-im/internal/models"
)

type MessageRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{db: db, collection: db.Collection(models.CollectionMessage)}
}

func (r *MessageRepository) Save(ctx context.Context, message *models.Message) (*models.Message, error) {
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	if message.ID.IsZero() {
		message.ID = primitive.NewObjectID()
	}

	var err error
	if message.Payload != nil {
		err = message.EncodePayload()
		if err != nil {
			return nil, err
		}
	}

	_, err = r.collection.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (r *MessageRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Message, error) {
	var message models.Message
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *MessageRepository) FindByClientMessageID(ctx context.Context, conversationID primitive.ObjectID, senderID string, clientMessageID string) (*models.Message, error) {
	var message models.Message
	err := r.collection.FindOne(ctx, bson.M{
		"conversation_id":   conversationID,
		"sender_id":         senderID,
		"client_message_id": clientMessageID,
	}).Decode(&message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

// FindMessagesByConversationID 根据会话ID查询消息
func (r *MessageRepository) FindMessagesByConversationID(ctx context.Context, conversationID primitive.ObjectID, beforeID *primitive.ObjectID, afterID *primitive.ObjectID, limit int64) ([]models.Message, error) {
	filter := bson.M{"conversation_id": conversationID}

	sort := bson.M{"created_at": -1}
	if beforeID != nil && !beforeID.IsZero() {
		filter["_id"] = bson.M{"$lt": beforeID}
	}
	if afterID != nil && !afterID.IsZero() {
		filter["_id"] = bson.M{"$gt": afterID}
		sort = bson.M{"created_at": 1}
	}
	opts := options.Find().SetSort(sort).SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
