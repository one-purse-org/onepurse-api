package model

import (
	"time"
)

type User struct {
	ID                  string              `bson:"_id, omitempty" json:"id,omitempty"`
	FullName            string              `bson:"full_name, omitempty" json:"full_name,omitempty"`
	UserName            string              `bson:"username, omitempty" json:"username,omitempty"`
	PhoneNumber         string              `bson:"phone_number, omitempty" json:"phone_number,omitempty"`
	Email               string              `bson:"email, omitempty" json:"email,omitempty"`
	TransactionPassword string              `bson:"transaction_password" json:"transaction_password,omitempty"`
	Wallet              map[string]Wallet   `bson:"wallet, omitempty" json:"wallet,omitempty"` // a map of wallet with the keys representing the currency and the value the wallet info for that currency
	PlaidAccessToken    string              `bson:"plaid_access_token, omitempty" json:"plaid_access_token,omitempty"`
	Location            string              `bson:"location, omitempty" json:"location,omitempty"`
	Nationality         string              `bson:"nationality, omitempty" json:"nationality,omitempty"`
	DateOfBirth         string              `bson:"date_of_birth, omitempty" json:"date_of_birth,omitempty"`
	Gender              string              `json:"gender,omitempty" bson:"gender, omitempty"`
	Avatar              string              `bson:"avatar, omitempty" json:"avatar,omitempty"`
	IDType              string              `bson:"id_type, omitempty" json:"id_type,omitempty"`
	IDNumber            string              `bson:"id_number, omitempty" json:"id_number,omitempty"`
	IDExpiryDate        string              `bson:"id_expiry_date, omitempty" json:"id_expiry_date,omitempty"`
	PreferredCurrency   []PreferredCurrency `bson:"preferred_currency, omitempty" json:"preferred_currency,omitempty"`
	IDImage             string              `bson:"id_image, omitempty" json:"id_image,omitempty"`
	IsIDVerified        bool                `bson:"is_id_verified, omitempty" json:"is_id_verified,omitempty"`
	CreatedAt           time.Time           `bson:"created_at, omitempty" json:"created_at,omitempty"`
	DeviceToken         string              `bson:"device_token, omitempty" json:"device_token,omitempty"`
	Active              bool                `bson:"active, omitempty" json:"active,omitempty"`
}

type UpdateKYCInfo struct {
	Location          string              `bson:"location, omitempty" json:"location,omitempty"`
	Nationality       string              `bson:"nationality, omitempty" json:"nationality,omitempty"`
	DateOfBirth       string              `bson:"date_of_birth, omitempty" json:"date_of_birth,omitempty"`
	Gender            string              `json:"gender,omitempty" bson:"gender, omitempty"`
	IDType            string              `bson:"id_type, omitempty" json:"id_type,omitempty"`
	IDNumber          string              `bson:"id_number, omitempty" json:"id_number,omitempty"`
	PreferredCurrency []PreferredCurrency `bson:"preferred_currency, omitempty" json:"preferred_currency,omitempty"`
	IDExpiryDate      string              `bson:"id_expiry_date, omitempty" json:"id_expiry_date,omitempty"`
	IDImage           string              `bson:"id_image, omitempty" json:"id_image,omitempty"`
}

type UpdateUserInfo struct {
	PhoneNumber string `bson:"phone_number, omitempty" json:"phone_number,omitempty"`
	Email       string `bson:"email, omitempty" json:"email,omitempty"`
}

type PreferredCurrency struct {
	Label string `bson:"slug" json:"label"`
	Slug  string `bson:"slug" json:"slug"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User         *User     `json:"user"`
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

type ChangeTransactionPasswordRequest struct {
	ProposedPassword string `json:"proposed_password"`
	OTP              string `json:"otp"`
}

type ChangePassword struct {
	AccessToken      string `json:"access_token"`
	PreviousPassword string `json:"previous_password"`
	ProposedPassword string `json:"proposed_password"`
}

type UpdateUsername struct {
	AccessToken       string `json:"access_token"`
	PreferredUsername string `json:"preferred_username"`
}

// UserAccount is the model for user bank account information
type UserAccount struct {
	ID        string    `bson:"_id" json:"id"`
	User      *User     `bson:"user" json:"user"`
	Name      string    `bson:"name" json:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
