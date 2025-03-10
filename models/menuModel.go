package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Menu struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `json:"name" bson:"name" validate:"required,min=2,max=100"`
	Category   string             `json:"category" bson:"category" validate:"required,min=2,max=50"`
	UniqueID   string             `json:"unique_id" bson:"unique_id"`
	Created_at time.Time          `json:"created_at" bson:"created_at"`
	Updated_at time.Time          `json:"updated_at" bson:"updated_at"`
	Menu_id    string             `json:"menu_id" bson:"menu_id"`
}
