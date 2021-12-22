package model

import "time"

type User struct {
	ID                  string                `bson:"_id" json:"id"`
	FullName            string                `bson:"full_name" json:"full_name"`
	UserName            string                `bson:"username" json:"username"`
	Email               string                `bson:"email" json:"email"`
	TransactionPassword string                `bson:"transaction_password" json:"transaction_password"`
	Wallet              map[string]UserWallet `bson:"wallet" json:"wallet"` // a map of wallet with the keys representing the currency and the value the wallet info for that currency
	PlaidAccessToken    string                `bson:"plaid_access_token" json:"plaid_access_token"`
	Location            string                `bson:"location" json:"location"`
	Nationality         string                `bson:"nationality" json:"nationality"`
	DateOfBirth         time.Time             `bson:"date_of_birth" json:"date_of_birth"`
	Gender              string                `json:"gender" bson:"gender"`
	Avatar              string                `bson:"avatar" json:"avatar"`
	IDType              string                `bson:"id_type" json:"id_type"`
	IDNumber            string                `bson:"id_number" json:"id_number"`
	IDExpiryDate        time.Time             `bson:"id_expiry_date" json:"id_expiry_date"`
	PreferredCurrency   []PreferredCurrency   `bson:"preferred_currency" json:"preferred_currency"`
	IDImage             string                `bson:"id_image" json:"id_image"`
	IsIDVerified        bool                  `bson:"is_id_verified" json:"is_id_verified"`
	CreatedAt           time.Time             `bson:"created_at" json:"created_at"`
	DeviceToken         string                `bson:"device_token" json:"device_token"`
	Active              bool                  `bson:"active" json:"active"`
}

type UserWallet struct {
	Currency         string  `bson:"currency" json:"currency"`                   // NGN, USD BSD
	AvailableBalance float32 `bson:"available_balance" json:"available_balance"` // Amount that can be withdrawn
	PendingBalance   float32 `bson:"pending_balance" json:"pending_balance"`     // Amount tied up in transactions
	TotalVolume      float32 `bson:"total_volume" json:"total_volume"`           // Total Amount transacted with this wallet. Might not be necessary
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

type ChangePassword struct {
	AccessToken      string `json:"access_token"`
	PreviousPassword string `json:"previous_password"`
	ProposedPassword string `json:"proposed_password"`
}

type UpdateUsername struct {
	AccessToken       string `json:"access_token"`
	PreferredUsername string `json:"preferred_username"`
}

type Notification struct {
	ID        string    `bson:"_id" json:"id"`
	Title     string    `bson:"title" json:"title"`
	Message   string    `bson:"message" json:"message"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Read      bool      `bson:"read" json:"read"`
}

// UserAccount is the model for user bank account information
type UserAccount struct {
	ID        string    `bson:"_id" json:"id"`
	User      *User     `bson:"user" json:"user"`
	Name      string    `bson:"name" json:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
