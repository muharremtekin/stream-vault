package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NotificationPreferences struct {
	ID        bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID    string            `bson:"userId" json:"userId"`
	Email     ChannelPreference `bson:"email" json:"email"`
	Push      ChannelPreference `bson:"push" json:"push"`
	InApp     ChannelPreference `bson:"inApp" json:"inApp"`
	UpdatedAt time.Time         `bson:"updatedAt" json:"updatedAt"`
}

type ChannelPreference struct {
	Enabled    bool     `bson:"enabled" json:"enabled"`
	Categories []string `bson:"categories,omitempty" json:"categories,omitempty"`
}

type PreferencesStore interface {
	Get(ctx context.Context, userID string) (*NotificationPreferences, error)
	Upsert(ctx context.Context, prefs *NotificationPreferences) error
}

type MongoPreferencesStore struct {
	collection *mongo.Collection
}

func NewMongoPreferencesStore(db *mongo.Database) *MongoPreferencesStore {
	return &MongoPreferencesStore{
		collection: db.Collection("notification_preferences"),
	}
}

func (s *MongoPreferencesStore) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := s.collection.Indexes().CreateOne(ctx, index)
	if err != nil {
		return fmt.Errorf("creating preferences index: %w", err)
	}
	return nil
}

func (s *MongoPreferencesStore) Get(ctx context.Context, userID string) (*NotificationPreferences, error) {
	filter := bson.D{{Key: "userId", Value: userID}}

	var prefs NotificationPreferences
	err := s.collection.FindOne(ctx, filter).Decode(&prefs)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("finding preferences for user %q: %w", userID, err)
	}
	return &prefs, nil
}

func (s *MongoPreferencesStore) Upsert(ctx context.Context, prefs *NotificationPreferences) error {
	prefs.UpdatedAt = time.Now().UTC()

	filter := bson.D{{Key: "userId", Value: prefs.UserID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "email", Value: prefs.Email},
		{Key: "push", Value: prefs.Push},
		{Key: "inApp", Value: prefs.InApp},
		{Key: "updatedAt", Value: prefs.UpdatedAt},
	}}}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := s.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("upserting preferences for user %q: %w", prefs.UserID, err)
	}
	return nil
}

func DefaultPreferences(userID string) *NotificationPreferences {
	return &NotificationPreferences{
		UserID: userID,
		Email:  ChannelPreference{Enabled: true},
		Push:   ChannelPreference{Enabled: true},
		InApp:  ChannelPreference{Enabled: true},
	}
}
