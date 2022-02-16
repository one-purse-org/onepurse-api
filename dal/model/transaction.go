package model

import "time"

type Withdrawal struct {
	ID           string       `bson:"_id" json:"id"`
	BaseAmount   float32      `bson:"amount" json:"amount"`
	BaseCurrency string       `bson:"currency" json:"currency"` // USD, NGN, BS
	UserAccount  *UserAccount `bson:"user_account" json:"user_account"`
	Status       string       `bson:"status" json:"status"`
	CreatedAt    time.Time    `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `bson:"updated_at" json:"updated_at"`
}

type Transfer struct {
	ID             string    `bson:"_id" json:"id"`
	UserID         string    `bson:"user_id" json:"user_id"`
	AgentID        string    `bson:"agent_id" json:"agent_id"`
	BaseAmount     float32   `bson:"base_amount" json:"base_amount"`
	BaseCurrency   string    `bson:"base_currency" json:"base_currency"`
	AmountSent     float32   `bson:"amount_sent" json:"amount_sent"`
	ConvCurrency   string    `bson:"conv_currency" json:"conv_currency"`
	PaymentChannel string    `bson:"payment_channel" json:"payment_channel"`
	AgentAccount   *Account  `bson:"agent_account" json:"agent_account"`
	UserReceipt    string    `bson:"user_receipt" json:"user_receipt"`
	AgentReceipt   string    `bson:"agent_receipt" json:"agent_receipt"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
	Status         string    `bson:"status" json:"status"` // pending, completed, cancelled
}

//OnePurseTransaction refers to transfer between one purse users
type OnePurseTransaction struct {
	ID        string    `bson:"_id" json:"id"`
	FromUser  *User     `bson:"from_user" json:"from_user"` // user initiating the one purse transaction
	ToUser    *User     `bson:"to_user" json:"to_user"`
	Amount    float32   `bson:"amount" json:"amount"`
	Currency  string    `bson:"currency" json:"currency"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	Status    string    `bson:"status" json:"status"`
	Type      string    `bson:"type" json:"type"` // can either be pay or request
}

type Deposit struct {
	ID             string    `bson:"_id" json:"id"`
	UserID         string    `bson:"user_id" json:"user_id"`
	AgentID        string    `bson:"agent_id" json:"agent_id"`
	BaseCurrency   string    `bson:"base_currency" json:"base_currency"`
	BaseAmount     float32   `bson:"amount" json:"amount"`
	PaymentChannel string    `bson:"payment_channel" json:"payment_channel"`
	UserReceipt    string    `bson:"user_receipt" json:"user_receipt"`
	AgentAccount   *Account  `bson:"agent_account" json:"agent_account"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
	Status         string    `bson:"status" json:"status"`
}

type Exchange struct {
	ID                       string    `bson:"_id" json:"id"`
	UserID                   string    `bson:"user" json:"user"`
	AgentID                  string    `bson:"agent_id" json:"agent_id"`
	BaseCurrency             string    `bson:"base_currency" json:"base_currency"`
	BaseAmount               float32   `bson:"base_amount" json:"base_amount"`
	ExchangeCurrency         string    `bson:"exchange_currency" json:"exchange_currency"`
	ExchangeAmount           float32   `bson:"exchange_amount" json:"exchange_amount"`
	IsCryptoExchange         bool      `bson:"is_crypto_exchange" json:"is_crypto_exchange"`
	BlockchainChannel        string    `bson:"blockchain_channel" json:"blockchain_channel"`
	CryptoWalletAddress      string    `bson:"crypto_wallet_address" json:"crypto_wallet_address"`
	PaymentChannel           string    `bson:"payment_channel" json:"payment_channel"`
	AgentAccount             *Account  `bson:"agent_account" json:"agent_account"`
	UserReceipt              string    `bson:"user_receipt" json:"user_receipt"`
	AgentReceipt             string    `bson:"agent_receipt" json:"agent_receipt"`
	UserReasonForCancelling  string    `bson:"reason_for_cancelling" json:"reason_for_cancelling"`
	AgentReasonForCancelling string    `bson:"agent_reason_for_cancelling" json:"agent_reason_for_cancelling"`
	ReasonForDispute         string    `bson:"reason_for_dispute" json:"reason_for_dispute"`
	Status                   string    `bson:"status" json:"status"`
	CreatedAt                time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt                time.Time `bson:"updated_at" json:"updated_at"`
}

type Wallet struct {
	Currency         string    `bson:"currency" json:"currency"`                   // NGN, USD BSD
	AvailableBalance float32   `bson:"available_balance" json:"available_balance"` // BaseAmount that can be withdrawn
	PendingBalance   float32   `bson:"pending_balance" json:"pending_balance"`     // BaseAmount tied up in transactions
	TotalVolume      float32   `bson:"total_volume" json:"total_volume"`           // Total BaseAmount transacted with this wallet. Might not be necessary
	IsActive         bool      `bson:"is_active" json:"is_active"`
	CreatedAt        time.Time `bson:"created_at" json:"created_at"` // Date the wallet was created
	UpdatedAt        time.Time `bson:"updated_at" json:"updated_at"`
}

type Account struct {
	ID            string `bson:"_id" json:"id"`
	AgentID       string `bson:"agent_id" json:"agent_id"`
	AccountName   string `bson:"account_name" json:"account_name"`
	AccountNumber string `bson:"account_number" json:"account_number"`
	BankName      string `bson:"bank_name" json:"bank_name"`
	IsAgent       bool   `bson:"is_agent" json:"is_agent"`
	IsUser        bool   `bson:"is_user" json:"is_user"`
}

type Rate struct {
	NGN map[string]float32 `json:"ngn" bson:"ngn"`
	USD map[string]float32 `json:"usd" bson:"usd"`
	BSD map[string]float32 `json:"bsd" bson:"bsd"`
	BTC map[string]float32 `json:"btc" bson:"btc"`
}

type AdminPayment struct {
	RecipientName string  `json:"recipient_name" bson:"recipient_name"`
	Category      string  `json:"category" bson:"category"`
	BaseCurrency  string  `json:"base_currency" bson:"base_currency"`
	BaseAmount    float32 `json:"base_amount" bson:"base_amount"`
	ConvCurrency  string  `json:"conv_currency" bson:"conv_currency"`
	Description   string  `json:"description" bson:"description"`
	Receipt       string  `json:"receipt" bson:"receipt"`
}

type PaymentCategory struct {
}
