package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type TGAccount struct {
	ClientID primitive.ObjectID `bson:"_id"`
	ChatID   int64              `bson:"chatID"`
}
