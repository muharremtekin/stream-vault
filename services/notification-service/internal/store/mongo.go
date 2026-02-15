package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var mongoTracer = otel.Tracer("notification-service/store")

type Notification struct {
	ID        bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID    string          `bson:"userId" json:"userId"`
	Type      string          `bson:"type" json:"type"`
	Category  string          `bson:"category" json:"category"`
	Title     string          `bson:"title" json:"title"`
	Body      string          `bson:"body" json:"body"`
	Icon      string          `bson:"icon,omitempty" json:"icon,omitempty"`
	Action    string          `bson:"action,omitempty" json:"action,omitempty"`
	Read      bool            `bson:"read" json:"read"`
	Channels  []ChannelStatus `bson:"channels" json:"channels"`
	CreatedAt time.Time       `bson:"createdAt" json:"createdAt"`
	ExpiresAt time.Time       `bson:"expiresAt" json:"expiresAt"`
}

type ChannelStatus struct {
	Channel string    `bson:"channel" json:"channel"`
	Status  string    `bson:"status" json:"status"`
	SentAt  time.Time `bson:"sentAt,omitempty" json:"sentAt,omitempty"`
}

type NotificationStore interface {
	Create(ctx context.Context, n *Notification) error
	ListByUserID(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) ([]Notification, int64, error)
	MarkAsRead(ctx context.Context, id string, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	CountUnread(ctx context.Context, userID string) (int64, error)
}

type MongoNotificationStore struct {
	collection *mongo.Collection
}

func NewMongoNotificationStore(db *mongo.Database) *MongoNotificationStore {
	return &MongoNotificationStore{
		collection: db.Collection("notifications"),
	}
}

func (s *MongoNotificationStore) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "read", Value: 1},
			},
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}
	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("creating notification indexes: %w", err)
	}
	return nil
}

func (s *MongoNotificationStore) Create(ctx context.Context, n *Notification) error {
	ctx, span := mongoTracer.Start(ctx, "mongodb.Create",
		trace.WithAttributes(
			attribute.String("db.system", "mongodb"),
			attribute.String("db.name", "streamvault_notifications"),
			attribute.String("db.operation", "insertOne"),
		),
	)
	defer span.End()

	now := time.Now().UTC()
	n.CreatedAt = now
	n.Read = false
	if n.ExpiresAt.IsZero() {
		n.ExpiresAt = now.Add(90 * 24 * time.Hour)
	}

	_, err := s.collection.InsertOne(ctx, n)
	if err != nil {
		return fmt.Errorf("inserting notification: %w", err)
	}
	return nil
}

func (s *MongoNotificationStore) ListByUserID(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) ([]Notification, int64, error) {
	ctx, span := mongoTracer.Start(ctx, "mongodb.ListByUserID",
		trace.WithAttributes(
			attribute.String("db.system", "mongodb"),
			attribute.String("db.name", "streamvault_notifications"),
			attribute.String("db.operation", "find"),
		),
	)
	defer span.End()

	filter := bson.D{{Key: "userId", Value: userID}}
	if unreadOnly {
		filter = append(filter, bson.E{Key: "read", Value: false})
	}

	totalCount, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("counting notifications: %w", err)
	}

	skip := int64((page - 1) * pageSize)
	findOpts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(pageSize))

	cursor, err := s.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("finding notifications: %w", err)
	}
	defer cursor.Close(ctx)

	var notifications []Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, 0, fmt.Errorf("decoding notifications: %w", err)
	}

	if notifications == nil {
		notifications = []Notification{}
	}

	return notifications, totalCount, nil
}

func (s *MongoNotificationStore) MarkAsRead(ctx context.Context, id string, userID string) error {
	ctx, span := mongoTracer.Start(ctx, "mongodb.MarkAsRead",
		trace.WithAttributes(
			attribute.String("db.system", "mongodb"),
			attribute.String("db.name", "streamvault_notifications"),
			attribute.String("db.operation", "updateOne"),
		),
	)
	defer span.End()

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid notification id %q: %w", id, err)
	}

	filter := bson.D{
		{Key: "_id", Value: objID},
		{Key: "userId", Value: userID},
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "read", Value: true}}}}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("marking notification as read: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("notification %q not found for user %q", id, userID)
	}
	return nil
}

func (s *MongoNotificationStore) MarkAllAsRead(ctx context.Context, userID string) error {
	ctx, span := mongoTracer.Start(ctx, "mongodb.MarkAllAsRead",
		trace.WithAttributes(
			attribute.String("db.system", "mongodb"),
			attribute.String("db.name", "streamvault_notifications"),
			attribute.String("db.operation", "updateMany"),
		),
	)
	defer span.End()

	filter := bson.D{
		{Key: "userId", Value: userID},
		{Key: "read", Value: false},
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "read", Value: true}}}}

	_, err := s.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("marking all notifications as read: %w", err)
	}
	return nil
}

func (s *MongoNotificationStore) CountUnread(ctx context.Context, userID string) (int64, error) {
	ctx, span := mongoTracer.Start(ctx, "mongodb.CountUnread",
		trace.WithAttributes(
			attribute.String("db.system", "mongodb"),
			attribute.String("db.name", "streamvault_notifications"),
			attribute.String("db.operation", "countDocuments"),
		),
	)
	defer span.End()

	filter := bson.D{
		{Key: "userId", Value: userID},
		{Key: "read", Value: false},
	}
	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("counting unread notifications: %w", err)
	}
	return count, nil
}
