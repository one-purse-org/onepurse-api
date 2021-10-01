package model

import "time"

type User struct {
	ID        string    `bson:"_id" json:"id"`
	FullName  string    `bson:"full_name" json:"full_name"`
	UserName  string    `bson:"user_name" json:"user_name"`
	Email     string    `bson:"email" json:"email"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Active    bool      `bson:"active" json:"active"`
}

type LoginRequest struct {
	Email    string
	Password string
}

type AuthResponse struct {
	ID           string
	Email        string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type RegistrationRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

type SignupResponse struct {
	IsConfirmed    bool
	DeliveryMedium string
	Destination    string
}

type VerificationRequest struct {
	Email string
	Code  string
}
