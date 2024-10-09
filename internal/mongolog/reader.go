package mongolog

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogEntry struct {
	ID        string            `bson:"_id" json:"id"`
	CreatedAt time.Time         `bson:"created_at" json:"-"` // retains the full time precision to ensure accurate and performant sorting
	Details   map[string]string `bson:"details" json:"details"`
	LogTime   string            `bson:"time" json:"logTime"`
}

func (l *logStore) ReadLogsPaginated(ctx context.Context, limit, offset int64) ([]LogEntry, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	if limit < 1 {
		limit = 10
	}

	findOptions := options.Find()
	findOptions.SetSkip(offset)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by date created descending

	cursor, err := l.collection.Find(ctxWithTimeout, bson.D{}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []LogEntry
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	return logs, nil
}
