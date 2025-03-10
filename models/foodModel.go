package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Food struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Food_id      string             `json:"food_id"`
	Name         *string            `json:"name" validate:"required,min=2,max=100"`
	Price        *float64           `json:"price" validate:"required"`
	Food_image   *string            `json:"food_image"`
	Menu_id      *string            `json:"menu_id" validate:"required"`
	Created_at   time.Time          `json:"created_at"`
	Updated_at   time.Time          `json:"updated_at"`
	UniqueFoodID string             `bson:"unique_food_id" json:"unique_food_id"` // NEW FIELD
}
