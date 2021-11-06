package services

import (
	"context"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/plaid/plaid-go/plaid"
	"strings"
)

type PLAID struct {
	config      *config.Config
	plaidClient *plaid.APIClient
}

func NewPlaidService(cfg *config.Config) (*PLAID, error) {
	conf := plaid.NewConfiguration()
	conf.AddDefaultHeader("PLAID-CLIENT-ID", cfg.PlaidClientId)
	conf.AddDefaultHeader("PLAID-SECRET", cfg.PlaidSecret)
	conf.UseEnvironment(plaid.Sandbox) // TODO: make this automatic based on config environment value

	p := PLAID{
		config:      cfg,
		plaidClient: plaid.NewAPIClient(conf),
	}

	return &p, nil
}

func (p PLAID) CreateLinkToken(r model.CreateTokenLinkRequest) (string, error) {
	countryCodes := convertCountryCode(strings.Split(p.config.PlaidCountryCodes, ","))
	products := convertProducts(strings.Split(r.PlaidProduct, ","))
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: r.UserId,
		LegalName:    &r.UserName,
		PhoneNumber:  &r.PhoneNumber,
		EmailAddress: &r.Email,
		DateOfBirth:  &r.DateOfBirth,
	}

	req := plaid.NewLinkTokenCreateRequest(p.config.PlaidClientName, "en", countryCodes, user)

	req.SetProducts(products)

	if p.config.PlaidRedirectUri != "" {
		req.SetRedirectUri(p.config.PlaidRedirectUri)
	}
	if r.PaymentInitiation != nil {
		req.SetPaymentInitiation(*r.PaymentInitiation)
	}

	res, _, err := p.plaidClient.PlaidApi.LinkTokenCreate(context.TODO()).LinkTokenCreateRequest(*req).Execute()
	if err != nil {
		return "", err
	}
	return res.GetLinkToken(), nil
}

func (p PLAID) GetAccessToken(publicToken string) (string, error) {
	res, _, err := p.plaidClient.PlaidApi.ItemPublicTokenExchange(context.TODO()).ItemPublicTokenExchangeRequest(
		*plaid.NewItemPublicTokenExchangeRequest(publicToken)).Execute()
	if err != nil {
		return "", err
	}
	accessToken := res.GetAccessToken()
	return accessToken, nil
}

func (p PLAID) GetAccountInfo(accessToken string) (*[]plaid.AccountBase, error) {
	res, _, err := p.plaidClient.PlaidApi.AccountsGet(context.TODO()).AccountsGetRequest(
		*plaid.NewAccountsGetRequest(accessToken)).Execute()
	if err != nil {
		return nil, err
	}
	account := res.GetAccounts()
	return &account, nil
}

func convertCountryCode(countryCodeStr []string) []plaid.CountryCode {
	var countryCodes []plaid.CountryCode

	for _, countryCode := range countryCodeStr {
		countryCodes = append(countryCodes, plaid.CountryCode(countryCode))
	}

	return countryCodes
}

func convertProducts(productStr []string) []plaid.Products {
	var products []plaid.Products

	for _, product := range productStr {
		products = append(products, plaid.Products(product))
	}

	return products
}
