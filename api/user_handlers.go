package api

import (
	"fmt"
	"github.com/aws/smithy-go"
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/helpers"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"github.com/lucsky/cuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
	"time"
)

func (a *API) UserRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)

	// Profile Routes
	router.Method("POST", "/change_password", Handler(a.changePassword))
	router.Method("POST", "/{userID}/create_username", Handler(a.createUserName))
	router.Method("PATCH", "/{userID}/create_transaction_password", Handler(a.createTransactionPassword))
	router.Method("PATCH", "/{userID}/update_kyc_information", Handler(a.updateKYCInformation))

	// Transaction Routes
	router.Method("POST", "/{userID}/transaction", Handler(a.createTransaction))
	router.Method("PUT", "/transaction/{transactionID}/", Handler(a.updateTransaction))
	router.Method("GET", "/{userID}/transaction", Handler(a.getTransaction))

	return router
}

// Profile

func (a *API) changePassword(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var password model.ChangePassword
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &password); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}
	password.AccessToken = strings.Split(r.Header.Get("Authorization"), " ")[1]

	if password.PreviousPassword == "" {
		return RespondWithError(nil, "Previous password is required", http.StatusBadRequest, &tracingContext)
	}

	if password.ProposedPassword == "" {
		return RespondWithError(nil, "Proposed password is required", http.StatusBadRequest, &tracingContext)
	}

	status, err := a.Deps.AWS.Cognito.ChangePassword(&password)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			switch ae.ErrorCode() {
			case "InvalidParameterException":
				return RespondWithError(err, "Invalid parameters provided", http.StatusBadRequest, &tracingContext)
			case "NotAuthorizedException":
				return RespondWithError(err, "Not authorized", http.StatusBadRequest, &tracingContext)
			case "ExpiredCodeException":
				return RespondWithError(err, "expired code used. request for a new one", http.StatusBadRequest, &tracingContext)
			case "CodeMismatchException":
				return RespondWithError(err, "Invalid code used", http.StatusBadRequest, &tracingContext)
			}
		}
	}

	response := map[string]interface{}{
		"status": status,
	}

	return &ServerResponse{Payload: response}
}

func (a *API) createTransactionPassword(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var user model.User
	userID := chi.URLParam(r, "userID")
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &user); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	if user.TransactionPassword == "" {
		return RespondWithError(nil, "Transaction password is required", http.StatusBadRequest, &tracingContext)
	}

	password, err := helpers.HashPassword(user.TransactionPassword)
	if err != nil {
		return RespondWithError(err, "Failed to hash password", http.StatusInternalServerError, &tracingContext)
	}
	err = a.Deps.DAL.UserDAL.UpdateUser(userID, bson.D{{"$set", bson.D{{"transaction_password", password}}}})
	if err != nil {
		return RespondWithError(err, "Failed to update transaction password", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"message": "transaction password updated",
	}

	return &ServerResponse{Payload: response}
}

func (a *API) createUserName(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var param model.UpdateUsername
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	userID := chi.URLParam(r, "userID")

	if err := decodeJSONBody(&tracingContext, r.Body, &param); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	if param.AccessToken == "" {
		return RespondWithError(nil, "access_token is required", http.StatusBadRequest, &tracingContext)

	}
	if param.PreferredUsername == "" {
		return RespondWithError(nil, "preferred_username is required", http.StatusBadRequest, &tracingContext)
	}

	err := a.Deps.AWS.Cognito.UpdateUsername(&param)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			switch ae.ErrorCode() {
			case "InvalidParameterException":
				return RespondWithError(err, "Invalid parameters provided", http.StatusBadRequest, &tracingContext)
			case "NotAuthorizedException":
				return RespondWithError(err, "Not authorized", http.StatusBadRequest, &tracingContext)
			case "ExpiredCodeException":
				return RespondWithError(err, "expired code used. request for a new one", http.StatusBadRequest, &tracingContext)
			case "CodeMismatchException":
				return RespondWithError(err, "Invalid code used", http.StatusBadRequest, &tracingContext)
			}
		}
	}

	err = a.Deps.DAL.UserDAL.UpdateUser(userID, bson.D{{"username", param.PreferredUsername}})
	if err != nil {
		return RespondWithError(err, "Failed to update user name", http.StatusInternalServerError, &tracingContext)

	}

	response := map[string]interface{}{
		"message": "transaction password updated",
	}

	return &ServerResponse{Payload: response}
}

func (a *API) updateKYCInformation(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var user model.User
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	userID := chi.URLParam(r, "userID")

	if err := decodeJSONBody(&tracingContext, r.Body, &user); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	if user.Location == "" {
		return RespondWithError(nil, "location is required", http.StatusBadRequest, &tracingContext)
	}
	if user.Nationality == "" {
		return RespondWithError(nil, "nationality is required", http.StatusBadRequest, &tracingContext)
	}
	if user.DateOfBirth.IsZero() {
		return RespondWithError(nil, "date_of_birth is required", http.StatusBadRequest, &tracingContext)
	}
	if user.Gender == "" {
		return RespondWithError(nil, "gender is required", http.StatusBadRequest, &tracingContext)
	}
	if user.IDType == "" {
		return RespondWithError(nil, "id_type is required", http.StatusBadRequest, &tracingContext)
	}
	if user.IDNumber == "" {
		return RespondWithError(nil, "id_number is required", http.StatusBadRequest, &tracingContext)
	}
	if user.IDExpiryDate.IsZero() {
		return RespondWithError(nil, "id_expiry_date is required", http.StatusBadRequest, &tracingContext)
	}
	if user.IDImage == "" {
		return RespondWithError(nil, "id_image is required", http.StatusBadRequest, &tracingContext)
	}

	val, err := bson.Marshal(user)
	if err != nil {
		return RespondWithError(err, "Failed to marshal to update value", http.StatusInternalServerError, &tracingContext)

	}

	var doc bson.D
	err = bson.Unmarshal(val, &doc)
	if err != nil {
		return RespondWithError(err, "Failed to unmarshal update value to doc", http.StatusInternalServerError, &tracingContext)
	}

	err = a.Deps.DAL.UserDAL.UpdateUser(userID, doc)
	if err != nil {
		return RespondWithError(err, "Failed to update KYC information", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"message": "KYC information successfully updated",
	}

	return &ServerResponse{Payload: response}
}

// Transaction

func (a *API) createTransaction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	transactionType := r.URL.Query().Get("transaction-type")
	userId := chi.URLParam(r, "userID")
	user, err := a.Deps.DAL.UserDAL.FindByID(userId)
	if err != nil {
		return RespondWithError(err, "Unable to get user information", http.StatusInternalServerError, &tracingContext)
	}
	switch transactionType {
	case "transfer":
		var transfer model.Transfer
		if err := decodeJSONBody(&tracingContext, r.Body, &transfer); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}
		if transfer.AgentAccount == nil {
			return RespondWithError(nil, "agent_account is required", http.StatusBadRequest, &tracingContext)
		}
		if transfer.BaseCurrency == "" || transfer.ConvCurrency == "" {
			return RespondWithError(nil, "base and conversion currency are required", http.StatusBadRequest, &tracingContext)
		}
		if transfer.BaseAmount == 0 || transfer.AmountSent == 0 {
			return RespondWithError(nil, "base and converted amount are required", http.StatusBadRequest, &tracingContext)
		}
		if transfer.PaymentChannel == "" {
			return RespondWithError(nil, "payment channel in use is required", http.StatusBadRequest, &tracingContext)
		}
		pass := helpers.DoSufficientFundsCheck(user, transfer.BaseAmount, transfer.BaseCurrency)
		if !pass {
			return RespondWithError(nil, "Insufficient Funds to initiate transfer. Please Top-up Wallet and try again", http.StatusBadRequest, &tracingContext)
		}
		transfer.Status = "created"
		transfer.ID = cuid.New()
		transfer.CreatedAt = time.Now()
		transfer.User = user

		err := a.Deps.DAL.TransactionDAL.CreateTransfer(&transfer)
		if err != nil {
			return RespondWithError(err, "Failed to initiate transfer. Please try again", http.StatusInternalServerError, &tracingContext)
		}
		response := map[string]interface{}{
			"message": "successfully initiated transfer",
		}
		return &ServerResponse{
			Payload: response,
		}
	case "withdraw":
		var withdrawal model.Withdrawal
		if err := decodeJSONBody(&tracingContext, r.Body, &withdrawal); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusBadRequest, &tracingContext)
		}
		if withdrawal.Currency == "" {
			return RespondWithError(nil, "withdrawal currency is required", http.StatusBadRequest, &tracingContext)
		}
		if withdrawal.Amount == 0 {
			return RespondWithError(nil, "Withdrawal amount cannot equal 0", http.StatusBadRequest, &tracingContext)
		}
		if withdrawal.UserAccount == nil {
			return RespondWithError(nil, "Destination account is required", http.StatusBadRequest, &tracingContext)
		}

		pass := helpers.DoSufficientFundsCheck(user, withdrawal.Amount, withdrawal.Currency)
		if !pass {
			return RespondWithError(nil, "Insufficient funds to withdraw from", http.StatusBadRequest, &tracingContext)
		}

		withdrawal.CreatedAt = time.Now()
		withdrawal.ID = cuid.New()
		withdrawal.Status = "created"
		err := a.Deps.DAL.TransactionDAL.CreateWithdrawal(&withdrawal)
		if err != nil {
			return RespondWithError(err, "Failed to initiate withdrawal. Please try again", http.StatusBadRequest, &tracingContext)
		}
		response := map[string]interface{}{
			"message": "successfully initiated withdrawal",
		}
		return &ServerResponse{
			Payload: response,
		}
	case "deposit":
		var deposit model.Deposit
		if err := decodeJSONBody(&tracingContext, r.Body, &deposit); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}
		if deposit.Currency == "" {
			return RespondWithError(nil, "deposit currency is required", http.StatusBadRequest, &tracingContext)
		}
		if deposit.Amount == 0 {
			return RespondWithError(nil, "deposit amount is required", http.StatusBadRequest, &tracingContext)
		}
		if deposit.PaymentChannel == "" {
			return RespondWithError(nil, "payment channel is required", http.StatusBadRequest, &tracingContext)
		}
		if deposit.AgentAccount == nil {
			return RespondWithError(nil, "agent account used is required", http.StatusBadRequest, &tracingContext)
		}
		deposit.Status = "created"
		deposit.ID = cuid.New()
		deposit.CreatedAt = time.Now()
		deposit.User = user

		err := a.Deps.DAL.TransactionDAL.CreateDeposit(&deposit)
		if err != nil {
			return RespondWithError(err, "Failed to initiate deposit, Please try again", http.StatusInternalServerError, &tracingContext)
		}
		response := map[string]interface{}{
			"message": "successfully initiated deposit",
		}
		return &ServerResponse{
			Payload: response,
		}
	case "exchange":

	default:
		return RespondWithError(nil, "This transaction type is not supported", http.StatusBadRequest, &tracingContext)
	}
	response := map[string]interface{}{
		"message": fmt.Sprintf("transaction type %s does not exist", transactionType),
	}
	return &ServerResponse{
		Payload: response,
	}
}

func (a *API) updateTransaction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	return &ServerResponse{
		Payload: nil,
	}
}

func (a *API) getTransaction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	return &ServerResponse{
		Payload: nil,
	}
}
