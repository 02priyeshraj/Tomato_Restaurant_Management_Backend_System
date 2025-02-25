package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Food struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       *string            `json:"name" bson:"name" validate:"required,min=2,max=100"`
	Price      *float64           `json:"price" bson:"price" validate:"required,gt=0"`
	Food_image *string            `json:"food_image" bson:"food_image" validate:"required"`
	Created_at time.Time          `json:"created_at" bson:"created_at"`
	Updated_at time.Time          `json:"updated_at" bson:"updated_at"`
	Food_id    string             `json:"food_id" bson:"food_id"`
	Menu_id    *string            `json:"menu_id" bson:"menu_id" validate:"required"`
}
