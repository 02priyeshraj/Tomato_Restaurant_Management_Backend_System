package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID         primitive.ObjectID `bson:"_id"`
	Text       string             `json:"text" bson:"text" validate:"required"`
	Title      string             `json:"title" bson:"title" validate:"required"`
	Created_at time.Time          `json:"created_at" bson:"created_at"`
	Updated_at time.Time          `json:"updated_at" bson:"updated_at"`
	Note_id    string             `json:"note_id" bson:"note_id"`
	Order_id   *string            `json:"order_id" bson:"order_id" validate:"required"`
}
