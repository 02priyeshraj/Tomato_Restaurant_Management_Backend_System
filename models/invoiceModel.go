package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Invoice struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Invoice_id     string             `json:"invoice_id"`
	Order_id       *string            `json:"order_id"`
	User_id        *string            `json:"user_id"`
	Payment_method *string            `json:"payment_method" validate:"eq=CARD|eq=CASH|eq="`
	Payment_status *string            `json:"payment_status" validate:"required,eq=PENDING|eq=PAID"`
	TotalPrice     float64            `json:"total_price" bson:"total_price"`
	Payment_date   time.Time          `json:"payment_date"`
	Created_at     time.Time          `json:"created_at"`
	Updated_at     time.Time          `json:"updated_at"`
}
