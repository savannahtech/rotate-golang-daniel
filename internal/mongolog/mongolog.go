package mongolog

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:generate mockgen -destination=../../mocks/mongolog/mock_mongolog.go -package=mongologmock -source=mongolog.go
type LogStore interface {
	Write(ctx context.Context, logDetail map[string]string) error
	Close(ctx context.Context) error
	ReadLogsPaginated(ctx context.Context, page, pageSize int64) ([]LogEntry, error)
}

type logStore struct {
	collection *mongo.Collection
}

func NewMongoLogStore(ctx context.Context, mongoURI, databaseName, collectionName string) (LogStore, error) {
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	db := client.Database(databaseName)
	collections, err := db.ListCollectionNames(ctx, map[string]interface{}{"name": collectionName})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	var collection *mongo.Collection
	if len(collections) == 0 {
		err := db.CreateCollection(ctx, collectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}

		collection = db.Collection(collectionName)

		err = createIndex(ctx, collection)
		if err != nil {
			return nil, fmt.Errorf("failed to created index on collection: %w", err)
		}

	} else {
		collection = db.Collection(collectionName)
	}

	return &logStore{
		collection: collection,
	}, nil
}

func createIndex(ctx context.Context, collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: -1}}, // -1 for descending order
		Options: options.Index().SetName("created_at_index"),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}

	return nil
}

func (l *logStore) Write(ctx context.Context, logDetail map[string]string) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	now := time.Now()
	logTime := now
	changeTime, err := strconv.ParseInt(logDetail["time"], 10, 64)
	if err == nil {
		logTime = time.Unix(changeTime, 0)
	}

	logEntry := LogEntry{
		ID:        uuid.NewString(),
		CreatedAt: now,
		Details:   logDetail,
		LogTime:   logTime.Format(time.RFC3339),
	}

	_, err = l.collection.InsertOne(ctxWithTimeout, logEntry)
	if err != nil {
		return fmt.Errorf("failed to insert log entry into mongolog store: %w", err)
	}

	return nil
}

func (l *logStore) Close(ctx context.Context) error {
	return l.collection.Database().Client().Disconnect(ctx)
}
