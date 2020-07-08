package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Customer struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty"`
	Username   string               `bson:"username, omitempty"`
	TelegramID int                  `bson:"telegramID, omitempty"`
	Bag        []primitive.ObjectID `bson:"bag, omitempty"`
}

type Position struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name, omitempty"`
	Category string             `bson:"category, omitempty"`
	Price    int                `bson:"price, omitempty"`
}
