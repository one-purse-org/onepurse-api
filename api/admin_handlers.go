package api

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"github.com/isongjosiah/work/onepurse-api/types"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
)

func (a *API) AdminRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)
	router.Method("POST", "/create_currency", Handler(a.createCurrency))
	router.Method("PATCH", "/{ID}/respond_to_transaction", Handler(a.transactionResponseAction))
	router.Method("PATCH", "/wallet_action", Handler(a.walletAction))
	router.Method("POST", "/agent", Handler(a.adminCreateAgent))
	return router
}

func (a *API) createCurrency(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var currency model.Currency
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &currency); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	if currency.Label == "" {
		return RespondWithError(nil, "label is required", http.StatusBadRequest, &tracingContext)
	}
	if currency.Slug == "" {
		return RespondWithError(nil, "slug is required", http.StatusBadRequest, &tracingContext)
	}

	err := a.Deps.DAL.CurrencyDAL.Add(context.TODO(), &currency)
	if err != nil {
		return RespondWithError(err, "Failed to add currency", http.StatusInternalServerError, &tracingContext)

	}

	response := map[string]interface{}{
		"message": "currency successfully added",
	}
	return &ServerResponse{Payload: response}
}

func (a *API) transactionResponseAction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	ID := chi.URLParam(r, "ID")
	res := r.URL.Query().Get("response")
	transactionType := r.URL.Query().Get("transaction-type")
	var err error
	query := bson.D{{"$set", bson.D{{"status", res}, {"updated_at", time.Now()}}}}

	switch transactionType {
	case types.TRANSFER:
		err = a.Deps.DAL.TransactionDAL.UpdateTransfer(context.TODO(), ID, query)

	case types.ONE_PURSE_TRANSACTION:
		err = a.Deps.DAL.TransactionDAL.UpdateOnePurseTransaction(context.TODO(), ID, query)

	case types.WITHDRAW:
		err = a.Deps.DAL.TransactionDAL.UpdateWithdrawal(context.TODO(), ID, query)

	case types.EXCHANGE:
		err = a.Deps.DAL.TransactionDAL.UpdateExchange(context.TODO(), ID, query)

	case types.DEPOSIT:
		err = a.Deps.DAL.TransactionDAL.UpdateDeposit(context.TODO(), ID, query)
	}

	if err != nil {
		return RespondWithError(err, "Failed to respond to transaction. Please try again", http.StatusInternalServerError, &tracingContext)
	}
	response := map[string]interface{}{
		"message": "successfully responded to transaction",
	}
	return &ServerResponse{
		Payload: response,
	}
}

func (a *API) walletAction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	ID := chi.URLParam(r, "ID")
	action := r.URL.Query().Get("action")
	var walletUpdate model.UpdateWallet
	var err error

	if err := decodeJSONBody(&tracingContext, r.Body, &walletUpdate); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	incQuery := bson.D{{"$inc", bson.D{
		{fmt.Sprintf("wallet.%s.available_deposit", walletUpdate.Currency), walletUpdate.Amount},
	}}}

	decQuery := bson.D{{"$inc", bson.D{
		{fmt.Sprintf("wallet.%s.available_deposit", walletUpdate.Currency), -walletUpdate.Amount},
	}}}

	if walletUpdate.Owner == types.USER {
		switch action {
		case types.DEPOSIT:
			err = a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), ID, incQuery)

		case types.WITHDRAW:
			err = a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), ID, decQuery)
		}
	} else if walletUpdate.Owner == types.AGENT {
		switch action {
		case types.DEPOSIT:
			err = a.Deps.DAL.AgentDAL.Update(context.TODO(), ID, incQuery)

		case types.WITHDRAW:
			err = a.Deps.DAL.AgentDAL.Update(context.TODO(), ID, decQuery)
		}
	}

	if err != nil {
		return RespondWithError(err, "unable to carry out wallet action", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"message": "wallet action performed successfully",
	}

	return &ServerResponse{
		Payload: response,
	}
}

func (a *API) adminCreateAgent(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	var agent model.Agent

	if agent.Name == "" {
		return RespondWithError(nil, "agent name is required", http.StatusBadRequest, &tracingContext)
	}
	if agent.Address == "" {
		return RespondWithError(nil, "agent address is required", http.StatusBadRequest, &tracingContext)
	}
	if agent.IDImage == "" || agent.IDNumber == "" || agent.IDType == "" {
		return RespondWithError(nil, "agent id information is required", http.StatusBadRequest, &tracingContext)
	}

	if err := decodeJSONBody(&tracingContext, r.Body, &agent); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusBadRequest, &tracingContext)
	}
	err := a.Deps.DAL.AgentDAL.Add(context.TODO(), &agent)
	if err != nil {
		return RespondWithError(err, "Unable to create agent", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"message": "agent created successfully",
	}

	return &ServerResponse{
		Payload: response,
	}
}
