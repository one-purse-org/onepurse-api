package api

import (
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

func (a *API) AccountRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)
	router.Method("POST", "/create_link_token", Handler(a.createLinkToken))
	router.Method("POST", "/{userID}/exchange_public_token", Handler(a.exchangePublicToken))
	router.Method("GET", "/{userID}/details", Handler(a.fetchAccountDetails))

	return router
}

func (a *API) createLinkToken(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	var tokenLinkRequest model.CreateTokenLinkRequest

	if err := decodeJSONBody(&tracingContext, r.Body, &tokenLinkRequest); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusBadRequest, &tracingContext)
	}
	token, err := a.Deps.PLAID.CreateLinkToken(tokenLinkRequest)
	if err != nil {
		return RespondWithError(err, "Could not create link token", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"token": token,
	}
	return &ServerResponse{Payload: response}
}

func (a *API) exchangePublicToken(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	var token struct {
		PublicToken string `json:"public_token"`
	}
	userID := chi.URLParam(r, "userID")

	if err := decodeJSONBody(&tracingContext, r.Body, &token); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusBadRequest, &tracingContext)
	}
	if token.PublicToken == "" {
		return RespondWithError(nil, "Public Token is required", http.StatusBadRequest, &tracingContext)
	}

	accessToken, err := a.Deps.PLAID.GetAccessToken(token.PublicToken)
	if err != nil {
		return RespondWithError(err, "could not exchange public token", http.StatusInternalServerError, &tracingContext)
	}
	err = a.Deps.DAL.UserDAL.UpdateUser(userID, bson.D{{"plaid_access_token", accessToken}})
	if err != nil {
		return RespondWithError(err, "could not save access token", http.StatusInternalServerError, &tracingContext)
	}
	response := map[string]interface{}{
		"access_token": accessToken,
	}
	return &ServerResponse{Payload: response}
}

func (a *API) fetchAccountDetails(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	userID := chi.URLParam(r, "userID")
	user, err := a.Deps.DAL.UserDAL.FindByID(userID)
	if err != nil {
		return RespondWithError(err, "could not fetch user", http.StatusInternalServerError, &tracingContext)
	}

	accountDetails, err := a.Deps.PLAID.GetAccountInfo(user.PlaidAccessToken)
	if err != nil {
		return RespondWithError(err, "could not fetch account details", http.StatusInternalServerError, &tracingContext)
	}
	return &ServerResponse{Payload: accountDetails}
	return nil
}
