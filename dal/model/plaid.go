package model

import "github.com/plaid/plaid-go/plaid"

type CreateTokenLinkRequest struct {
	PaymentInitiation *plaid.LinkTokenCreateRequestPaymentInitiation
	UserId            string
	UserName          string
	PhoneNumber       string
	Email             string
	DateOfBirth       string
	PlaidProduct      string
}
