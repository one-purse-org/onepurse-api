package model

import "github.com/plaid/plaid-go/plaid"

type CreateTokenLinkRequest struct {
	PaymentInitiation *plaid.LinkTokenCreateRequestPaymentInitiation
	UserId            string `json:"user_id"`
	UserName          string
	PhoneNumber       string
	Email             string
	DateOfBirth       string
	PlaidProduct      string
}
