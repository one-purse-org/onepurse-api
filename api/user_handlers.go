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
	router.Method("GET", "/transaction/{transactionID}/get_agent", Handler(a.getAgentForTransaction))

	// Wallet Routes
	router.Method("POST", "/{userID}/wallet", Handler(a.createWallet))
	router.Method("GET", "/{userID}/wallet", Handler(a.getWalletTransaction))
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

	doc, err := helpers.MarshalStructToBSONDoc(user)
	if err != nil {
		return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
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
		if withdrawal.BaseCurrency == "" {
			return RespondWithError(nil, "withdrawal currency is required", http.StatusBadRequest, &tracingContext)
		}
		if withdrawal.BaseAmount == 0 {
			return RespondWithError(nil, "Withdrawal amount cannot equal 0", http.StatusBadRequest, &tracingContext)
		}
		if withdrawal.UserAccount == nil {
			return RespondWithError(nil, "Destination account is required", http.StatusBadRequest, &tracingContext)
		}

		pass := helpers.DoSufficientFundsCheck(user, withdrawal.BaseAmount, withdrawal.BaseCurrency)
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
		if deposit.BaseCurrency == "" {
			return RespondWithError(nil, "deposit currency is required", http.StatusBadRequest, &tracingContext)
		}
		if deposit.BaseAmount == 0 {
			return RespondWithError(nil, "deposit amount is required", http.StatusBadRequest, &tracingContext)
		}
		if deposit.PaymentChannel == "" {
			return RespondWithError(nil, "payment channel is required", http.StatusBadRequest, &tracingContext)
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
		var exchange model.Exchange
		if err := decodeJSONBody(&tracingContext, r.Body, &exchange); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusBadRequest, &tracingContext)
		}
		if exchange.BaseAmount == 0 || exchange.BaseCurrency == "" {
			return RespondWithError(nil, "base amount and currency is required", http.StatusBadRequest, &tracingContext)
		}
		if exchange.ExchangeAmount == 0 || exchange.ExchangeCurrency == "" {
			return RespondWithError(nil, "exchange amount and currency is required", http.StatusBadRequest, &tracingContext)
		}
		if exchange.IsCryptoExchange == false && (exchange.AgentAccount == nil || exchange.PaymentChannel == "") {
			return RespondWithError(nil, "agent account and payment channel is required", http.StatusBadRequest, &tracingContext)
		}
		if exchange.IsCryptoExchange == true && (exchange.BlockchainChannel == "" || exchange.CryptoWalletAddress == "") {
			return RespondWithError(nil, "crypto information is not provided", http.StatusBadRequest, &tracingContext)
		}

		pass := helpers.DoSufficientFundsCheck(user, exchange.BaseAmount, exchange.BaseCurrency)
		if !pass {
			return RespondWithError(nil, "insufficient funds to transfer from. Top-up Wallet", http.StatusBadRequest, &tracingContext)
		}

		exchange.CreatedAt = time.Now()
		exchange.ID = cuid.New()
		exchange.Status = "initiated"
		exchange.User = user
		err := a.Deps.DAL.TransactionDAL.CreateExchange(&exchange)
		if err != nil {
			return RespondWithError(err, "Failed to initiate transaction. Please try again", http.StatusBadRequest, &tracingContext)
		}
		response := map[string]interface{}{
			"message": "successfully initiated exchange",
		}
		return &ServerResponse{
			Payload: response,
		}

	default:
		return RespondWithError(nil, "This transaction type is not supported", http.StatusBadRequest, &tracingContext)
	}
}

func (a *API) updateTransaction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	transactionType := r.URL.Query().Get("transaction-type")
	transactionId := chi.URLParam(r, "transactionID")

	switch transactionType {
	case "transfer":
		var transfer model.Transfer

		if err := decodeJSONBody(&tracingContext, r.Body, &transfer); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		doc, err := helpers.MarshalStructToBSONDoc(transfer)
		if err != nil {
			return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
		}

		tempTransfer, err := a.Deps.DAL.TransactionDAL.GetTransferByID(transactionId)
		if err != nil {
			return RespondWithError(err, "error fetching transfer information", http.StatusBadRequest, &tracingContext)
		}
		if tempTransfer.Status == "completed" {
			// No update should be undertaken
			return RespondWithError(nil, "cannot update a completed transaction", http.StatusForbidden, &tracingContext)
		}

		transfer.UpdatedAt = time.Now()
		err = a.Deps.DAL.TransactionDAL.UpdateTransfer(transactionId, doc)
		if err != nil {
			return RespondWithError(err, "unable to update transfer information", http.StatusInternalServerError, &tracingContext)
		}

		response := map[string]interface{}{
			"message": "transfer information updated successfully",
		}
		return &ServerResponse{
			Payload: response,
		}
	case "withdraw":
		var withdrawal model.Withdrawal

		if err := decodeJSONBody(&tracingContext, r.Body, &withdrawal); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		doc, err := helpers.MarshalStructToBSONDoc(withdrawal)
		if err != nil {
			return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
		}

		tempWithdrawal, err := a.Deps.DAL.TransactionDAL.GetWithdrawalByID(transactionId)
		if err != nil {
			return RespondWithError(err, "error fetching withdrawal information", http.StatusBadRequest, &tracingContext)
		}
		if tempWithdrawal.Status == "completed" {
			return RespondWithError(nil, "cannot update a completed transaction", http.StatusForbidden, &tracingContext)

		}

		withdrawal.UpdatedAt = time.Now()
		err = a.Deps.DAL.TransactionDAL.UpdateWithdrawal(transactionId, doc)
		if err != nil {
			return RespondWithError(err, "unable to update withdrawal information", http.StatusInternalServerError, &tracingContext)
		}

		response := map[string]interface{}{
			"message": "withdrawal information updated successfully",
		}
		return &ServerResponse{
			Payload: response,
		}
	case "deposit":
		var deposit model.Deposit
		if err := decodeJSONBody(&tracingContext, r.Body, &deposit); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		doc, err := helpers.MarshalStructToBSONDoc(deposit)
		if err != nil {
			return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
		}

		tempDeposit, err := a.Deps.DAL.TransactionDAL.GetDepositByID(transactionId)
		if err != nil {
			return RespondWithError(err, "error fetching deposit information", http.StatusBadRequest, &tracingContext)
		}
		if tempDeposit.Status == "completed" {
			return RespondWithError(nil, "cannot update a completed transaction", http.StatusForbidden, &tracingContext)
		}

		deposit.UpdatedAt = time.Now()
		err = a.Deps.DAL.TransactionDAL.UpdateDeposit(transactionId, doc)
		if err != nil {
			return RespondWithError(err, "unable to update deposit information", http.StatusInternalServerError, &tracingContext)
		}

		response := map[string]interface{}{
			"message": "update information updated successfully",
		}
		return &ServerResponse{
			Payload: response,
		}
	case "exchange":
		var exchange model.Exchange

		if err := decodeJSONBody(&tracingContext, r.Body, &exchange); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		doc, err := helpers.MarshalStructToBSONDoc(exchange)
		if err != nil {
			return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
		}
		tempExchange, err := a.Deps.DAL.TransactionDAL.GetExchangeByID(transactionId)
		if err != nil {
			return RespondWithError(err, "error fetching exchange information", http.StatusForbidden, &tracingContext)
		}
		if tempExchange.Status == "completed" {
			return RespondWithError(nil, "cannot update a completed transaction", http.StatusForbidden, &tracingContext)
		}

		exchange.UpdatedAt = time.Now()
		err = a.Deps.DAL.TransactionDAL.UpdateExchange(transactionId, doc)
		if err != nil {
			return RespondWithError(err, "unable to update exchange information", http.StatusInternalServerError, &tracingContext)
		}

		response := map[string]interface{}{
			"message": "exchange information updated successfully",
		}
		return &ServerResponse{
			Payload: response,
		}
	default:
		return RespondWithError(nil, "This transaction type is not supported", http.StatusInternalServerError, &tracingContext)
	}
}

func (a *API) getTransaction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	transactionType := r.URL.Query().Get("transaction-type")
	userId := chi.URLParam(r, "userID")
	query := bson.D{{"user._id", userId}}

	switch transactionType {
	case "all":
		var response map[string]interface{}
		transfers, err := a.Deps.DAL.TransactionDAL.FetchTransfers(query)
		if err != nil {
			return RespondWithError(err, "unable to fetch transfers", http.StatusInternalServerError, &tracingContext)
		}
		withdraws, err := a.Deps.DAL.TransactionDAL.FetchWithdrawals(query)
		if err != nil {
			return RespondWithError(err, "unable to fetch withdraws", http.StatusInternalServerError, &tracingContext)
		}
		deposits, err := a.Deps.DAL.TransactionDAL.FetchDeposits(query)
		if err != nil {
			return RespondWithError(err, "unable to fetch deposits", http.StatusInternalServerError, &tracingContext)
		}
		exchanges, err := a.Deps.DAL.TransactionDAL.FetchExchanges(query)
		if err != nil {
			return RespondWithError(err, "unable to fetch exchanges", http.StatusInternalServerError, &tracingContext)
		}

		response["transfer"] = transfers
		response["withdraws"] = withdraws
		response["deposits"] = deposits
		response["exchanges"] = exchanges

		return &ServerResponse{
			Payload: response,
		}
	case "transfers":
		transfers, err := a.Deps.DAL.TransactionDAL.FetchTransfers(query)
		if err != nil {
			return RespondWithError(err, "unable to fetch transfers", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: transfers,
		}
	case "withdraws":
		withdraws, err := a.Deps.DAL.TransactionDAL.FetchWithdrawals(query)
		if err != nil {
			return RespondWithError(err, "unable to fetch withdraws", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: withdraws,
		}
	case "deposit":
		deposits, err := a.Deps.DAL.TransactionDAL.FetchDeposits(query)
		if err != nil {
			return RespondWithError(err, "unable to fetch deposits", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: deposits,
		}
	case "exchange":
		exchanges, err := a.Deps.DAL.TransactionDAL.FetchExchanges(query)
		if err != nil {
			return RespondWithError(err, "unable to fetch exchanges", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: exchanges,
		}
	default:
		return RespondWithError(nil, "This transaction type is not supported", http.StatusInternalServerError, &tracingContext)
	}
}

func (a *API) getAgentForTransaction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	transactionId := chi.URLParam(r, "transactionID")
	transactionType := r.URL.Query().Get("transaction-type")
	//baseCurrency := r.URL.Query().Get("base-currency")
	//bAmount := r.URL.Query().Get("base-amount")
	//amount, err := strconv.ParseFloat(bAmount, 32)
	//if err != nil {
	//	return RespondWithError(err, "unable to parse base amount", http.StatusBadRequest, &tracingContext)
	//}
	//baseAmount := float32(amount)
	//
	//if transactionType == "" {
	//	return RespondWithError(nil, "transaction type is required", http.StatusBadRequest, &tracingContext)
	//}
	//if baseCurrency == "" || baseAmount == 0 {
	//	return RespondWithError(nil, "transaction base currency is required", http.StatusBadRequest, &tracingContext)
	//}

	switch transactionType {
	case "transfer":
		transfer, err := a.Deps.DAL.TransactionDAL.GetTransferByID(transactionId)
		if err != nil {
			return RespondWithError(err, "could not fetch exchange information", http.StatusInternalServerError, &tracingContext)
		}
		query := bson.D{{"$gte", bson.D{{
			fmt.Sprintf("wallet.available_balance"), transfer.BaseAmount}}},
			{"wallet.currency", transfer.BaseCurrency}}
		agent, err := a.Deps.DAL.AgentDAL.FindOne(query)
		if err != nil {
			return RespondWithError(err, "unable to find an agent for your transaction right now", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: agent,
		}

	case "exchange":
		exchange, err := a.Deps.DAL.TransactionDAL.GetExchangeByID(transactionId)
		if err != nil {
			return RespondWithError(err, "could not fetch exchange information", http.StatusInternalServerError, &tracingContext)
		}
		query := bson.D{{"$gte", bson.D{{
			fmt.Sprintf("wallet.available_balance"), exchange.BaseAmount}}},
			{"wallet.currency", exchange.BaseCurrency}}
		agent, err := a.Deps.DAL.AgentDAL.FindOne(query)
		if err != nil {
			return RespondWithError(err, "unable to find an agent for your transaction right now", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: agent,
		}

	case "deposit":
		deposit, err := a.Deps.DAL.TransactionDAL.GetDepositByID(transactionId)
		if err != nil {
			return RespondWithError(err, "could not fetch deposit information", http.StatusInternalServerError, &tracingContext)
		}
		query := bson.D{{"$gte", bson.D{{
			fmt.Sprintf("wallet.available_balance"), deposit.BaseAmount}}},
			{"wallet.currency", deposit.BaseCurrency}}
		agent, err := a.Deps.DAL.AgentDAL.FindOne(query)
		if err != nil {
			return RespondWithError(err, "unable to find an agent for your transaction right now", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: agent,
		}
	}

	return &ServerResponse{
		Payload: nil,
	}
}

// Wallet

func (a *API) createWallet(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	walletType := r.URL.Query().Get("wallet-type")
	userId := chi.URLParam(r, "userID")

	if walletType == "" {
		return RespondWithError(nil, "wallet type is required", http.StatusBadRequest, &tracingContext)
	}

	wallet := &model.UserWallet{
		Currency:         walletType,
		AvailableBalance: 0,
		PendingBalance:   0,
		TotalVolume:      0,
		CreatedAt:        time.Now(),
	}
	doc, err := helpers.MarshalStructToBSONDoc(wallet)
	if err != nil {
		return RespondWithError(err, "unable to marshall to mongo document", http.StatusInternalServerError, &tracingContext)
	}
	err = a.Deps.DAL.UserDAL.UpdateUser(userId, bson.D{{"$set", bson.D{{
		fmt.Sprintf("wallet.%s", walletType), doc}}}})
	if err != nil {
		return RespondWithError(err, "unable to create wallet", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"message": fmt.Sprintf("successfully created %s wallet", walletType),
	}

	return &ServerResponse{
		Payload: response,
	}
}

func (a *API) getWalletTransaction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	walletType := r.URL.Query().Get("wallet-type")
	userId := chi.URLParam(r, "userID")
	var query = bson.D{{"base_currency", walletType}, {"user_id", userId}}

	var response map[string]interface{}
	transfers, err := a.Deps.DAL.TransactionDAL.FetchTransfers(query)
	if err != nil {
		return RespondWithError(err, "unable to fetch transfers", http.StatusInternalServerError, &tracingContext)
	}
	withdraws, err := a.Deps.DAL.TransactionDAL.FetchWithdrawals(query)
	if err != nil {
		return RespondWithError(err, "unable to fetch withdraws", http.StatusInternalServerError, &tracingContext)
	}
	deposits, err := a.Deps.DAL.TransactionDAL.FetchDeposits(query)
	if err != nil {
		return RespondWithError(err, "unable to fetch deposits", http.StatusInternalServerError, &tracingContext)
	}
	exchanges, err := a.Deps.DAL.TransactionDAL.FetchExchanges(query)
	if err != nil {
		return RespondWithError(err, "unable to fetch exchanges", http.StatusInternalServerError, &tracingContext)
	}

	response["transfer"] = transfers
	response["withdraws"] = withdraws
	response["deposits"] = deposits
	response["exchanges"] = exchanges

	return &ServerResponse{
		Payload: response,
	}

}
