package repository

import (
	"context"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/entities"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Repository struct {
	collection *mongo.Collection
	tracer     trace.Tracer
}

const _telegramCollectionName = "Telegram"

var errRecordExists = errors.New("error record exists")

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{
		collection: database.Collection(_telegramCollectionName),
		tracer:     otel.GetTracerProvider().Tracer("repo"),
	}
}

func (r Repository) Create(ctx context.Context, client entities.TGAccount) error {
	ctx, span := r.tracer.Start(ctx, "repo.Create")
	defer span.End()

	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": client.ClientID})
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "collection.CountDocuments")
	}

	if count != 0 {
		return errRecordExists
	}

	if _, err = r.collection.InsertOne(ctx, client); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "collection.InsertOne")
	}

	return nil
}

func (r Repository) Delete(ctx context.Context, clientID string) error {
	ctx, span := r.tracer.Start(ctx, "repo.Delete")
	defer span.End()

	castedID, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return errors.Wrap(err, "primitive.ObjectIDFromHex")
	}

	filter := bson.M{
		"_id": castedID,
	}

	if _, err = r.collection.DeleteOne(ctx, filter); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "collection.DeleteOne")
	}

	return nil
}

func (r Repository) GetChatIDByClientID(ctx context.Context, clientID string) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "repo.GetChatIDByClientID")
	defer span.End()

	castedID, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return 0, err
	}

	filter := bson.M{
		"_id": castedID,
	}

	var client entities.TGAccount
	if err = r.collection.FindOne(ctx, filter).Decode(&client); err != nil {
		return 0, err
	}

	return client.ChatID, nil
}
