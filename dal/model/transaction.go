package model

import "time"

type Withdrawal struct {
	ID          string       `bson:"_id" json:"id"`
	Amount      float32      `bson:"amount" json:"amount"`
	Currency    string       `bson:"currency" json:"currency"` // USD, NGN, BS
	UserAccount *UserAccount `bson:"user_account" json:"user_account"`
	Status      string       `bson:"status" json:"status"`
	CreatedAt   time.Time    `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `bson:"updated_at" json:"updated_at"`
}

type Transfer struct {
	ID             string        `bson:"_id" json:"id"`
	User           *User         `bson:"user" json:"user"`
	BaseAmount     float32       `bson:"base_amount" json:"base_amount"`
	BaseCurrency   string        `bson:"base_currency" json:"base_currency"`
	AmountSent     float32       `bson:"amount_sent" json:"amount_sent"`
	ConvCurrency   string        `bson:"conv_currency" json:"conv_currency"`
	PaymentChannel string        `bson:"payment_channel" json:"payment_channel"`
	AgentAccount   *AgentAccount `bson:"agent_account" json:"agent_account"`
	UserReceipt    string        `bson:"user_receipt" json:"user_receipt"`
	AgentReceipt   string        `bson:"agent_receipt" json:"agent_receipt"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updated_at"`
	Status         string        `bson:"status" json:"status"` // pending, completed, cancelled
}

type Deposit struct {
	ID             string        `bson:"_id" json:"id"`
	User           *User         `bson:"user" json:"user"`
	Currency       string        `bson:"currency" json:"currency"`
	Amount         float32       `bson:"amount" json:"amount"`
	PaymentChannel string        `bson:"payment_channel" json:"payment_channel"`
	UserReceipt    string        `bson:"user_receipt" json:"user_receipt"`
	AgentAccount   *AgentAccount `bson:"agent_account" json:"agent_account"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updated_at"`
	Status         string        `bson:"status" json:"status"`
}

type Exchange struct {
	ID                       string        `bson:"_id" json:"id"`
	BaseCurrency             string        `bson:"base_currency" json:"base_currency"`
	BaseAmount               float32       `bson:"base_amount" json:"base_amount"`
	ExchangeCurrency         string        `bson:"exchange_currency" json:"exchange_currency"`
	ExchangeAmount           float32       `bson:"exchange_amount" json:"exchange_amount"`
	IsCryptoExchange         bool          `bson:"is_crypto_exchange" json:"is_crypto_exchange"`
	BlockchainChannel        string        `bson:"blockchain_channel" json:"blockchain_channel"`
	CryptoWalletAddress      string        `bson:"crypto_wallet_address" json:"crypto_wallet_address"`
	PaymentChannel           string        `bson:"payment_channel" json:"payment_channel"`
	AgentAccount             *AgentAccount `bson:"agent_account" json:"agent_account"`
	UserReceipt              string        `bson:"user_receipt" json:"user_receipt"`
	AgentReceipt             string        `bson:"agent_receipt" json:"agent_receipt"`
	UserReasonForCancelling  string        `bson:"reason_for_cancelling" json:"reason_for_cancelling"`
	AgentReasonForCancelling string        `bson:"agent_reason_for_cancelling" json:"agent_reason_for_cancelling"`
	ReasonForDispute         string        `bson:"reason_for_dispute" json:"reason_for_dispute"`
	Status                   string        `bson:"status" json:"status"`
	CreatedAt                time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt                time.Time     `bson:"updated_at" json:"updated_at"`
}
