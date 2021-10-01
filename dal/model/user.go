package model

import "time"

type User struct {
	ID        string    `bson:"_id" json:"id"`
	FullName  string    `bson:"full_name" json:"full_name"`
	UserName  string    `bson:"user_name" json:"user_name"`
	Email     string    `bson:"email" json:"email"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Active    bool      `bson:"active" json:"active"`
	Phone     string    `bson:"phone" json:"phone"`
}

type LoginRequest struct {
	UserName string
	Password string
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type RegistrationRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

type SignupResponse struct {
	IsConfirmed    bool   `json:"is_confirmed"`
	DeliveryMedium string `json:"delivery_medium"`
	Destination    string `json:"destination"`
}

type VerificationRequest struct {
	UserName string `json:"user_name"`
	Code     string `json:"code"`
}

type ResendConfirmationCodeRequest struct {
	UserName string
}
