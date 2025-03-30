package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Items         map[string]int     `bson:"items" json:"items" validate:"required"` // key: food_id, value: quantity
	TotalPrice    float64            `bson:"total_price" json:"total_price" validate:"required,gt=0"`
	Created_at    time.Time          `bson:"created_at" json:"created_at"`
	Updated_at    time.Time          `bson:"updated_at" json:"updated_at"`
	Order_item_id string             `bson:"order_item_id" json:"order_item_id"`
	Order_id      string             `bson:"order_id" json:"order_id" validate:"required"`
	Table_id      string             `bson:"table_id" json:"table_id" validate:"required"`
}
