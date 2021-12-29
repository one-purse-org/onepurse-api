package model

import "time"

type UserNotification struct {
	ID        string      `bson:"_id" json:"id"`
	UserID    string      `bson:"user_id" json:"user_id"`
	Title     string      `bson:"title" json:"title"`
	Message   string      `bson:"message" json:"message"`
	InfoType  string      `bson:"info_type" json:"info_type"`
	InfoData  interface{} `bson:"info_data" json:"info_data"`
	CreatedAt time.Time   `bson:"created_at" json:"created_at"`
	Read      bool        `bson:"read" json:"read"`
}
