package model

import "time"

type User struct {
	ID                  string    `bson:"_id" json:"id"`
	FullName            string    `bson:"full_name" json:"full_name"`
	UserName            string    `bson:"user_name" json:"user_name"`
	Email               string    `bson:"email" json:"email"`
	TransactionPassword string    `bson:"transaction_password" json:"transaction_password"`
	PlaidAccessToken    string    `bson:"plaid_access_token" json:"plaid_access_token"`
	Avatar              string    `bson:"avatar" json:"avatar"`
	IDType              string    `bson:"id_type" json:"id_type"`
	IDImage             string    `bson:"id_image" json:"id_image"`
	IsIDVerified        bool      `bson:"is_id_verified" json:"is_id_verified"`
	CreatedAt           time.Time `bson:"created_at" json:"created_at"`
	Active              bool      `bson:"active" json:"active"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
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
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ConfirmForgotPasswordRequest struct {
	Username         string `json:"username"`
	Code             string `json:"code"`
	ProposedPassword string `json:"proposed_password"`
}

type ChangePassword struct {
	AccessToken      string `json:"access_token"`
	PreviousPassword string `json:"previous_password"`
	ProposedPassword string `json:"proposed_password"`
}
