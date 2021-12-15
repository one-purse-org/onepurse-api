package model

import "time"

type Withdrawal struct {
	ID          string       `bson:"_id" json:"id"`
	Amount      float32      `bson:"amount" json:"amount"`
	Currency    string       `bson:"currency" json:"currency"` // USD, NGN, BS
	UserAccount *UserAccount `bson:"user_account" json:"user_account"`
	Status      string       `bson:"status" json:"status"`
}

type Transfer struct {
	ID             string        `bson:"_id" json:"id"`
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
