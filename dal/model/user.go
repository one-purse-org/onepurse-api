package model

import "time"

type User struct {
	ID        string    `bson:"id" json:"id"`
	FirstName string    `bson:"first_name" json:"first_name"`
	LastName  string    `bson:"last_name" json:"last_name"`
	UserName  string    `bson:"user_name" json:"user_name"`
	Email     string    `bson:"email" json:"email"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Active    bool      `bson:"active" json:"active"`
}
