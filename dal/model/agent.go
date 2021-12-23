package model

import "time"

type Agent struct {
	ID string `bson:"id" json:"id"`
}

type AgentAccount struct {
	ID     string      `bson:"_id" json:"id"`
	Agent  *Agent      `bson:"agent" json:"agent"`
	Name   string      `bson:"name" json:"name"`
	Wallet AgentWallet `bson:"wallet" json:"wallet"`
}

type AgentWallet struct {
	Currency         string    `bson:"currency" json:"currency"`                   // NGN, USD BSD
	AvailableBalance float32   `bson:"available_balance" json:"available_balance"` // BaseAmount that can be withdrawn
	PendingBalance   float32   `bson:"pending_balance" json:"pending_balance"`     // BaseAmount tied up in transactions
	TotalVolume      float32   `bson:"total_volume" json:"total_volume"`           // Total BaseAmount transacted with this wallet. Might not be necessary
	CreatedAt        time.Time `bson:"created_at" json:"created_at"`               // Date the wallet was created
}
