package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Menu struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `json:"name" bson:"name" validate:"required,min=2,max=100"`
	Category   string             `json:"category" bson:"category" validate:"required,min=2,max=50"`
	Start_Date *time.Time         `json:"start_date" bson:"start_date" validate:"omitempty,ltfield=End_Date"`
	End_Date   *time.Time         `json:"end_date" bson:"end_date" validate:"omitempty,gtfield=Start_Date"`
	Created_at time.Time          `json:"created_at" bson:"created_at"`
	Updated_at time.Time          `json:"updated_at" bson:"updated_at"`
	Menu_id    string             `json:"menu_id" bson:"menu_id"`
}
