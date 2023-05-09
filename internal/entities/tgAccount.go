package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

type TGAccount struct {
	ClientID primitive.ObjectID `bson:"_id" json:"clientID"`
	ChatID   int64              `bson:"chatID" json:"chatID"`
}
