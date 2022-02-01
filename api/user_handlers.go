package api

import (
	"context"
	"fmt"
	"github.com/aws/smithy-go"
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/helpers"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"github.com/isongjosiah/work/onepurse-api/types"
	"github.com/lucsky/cuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strings"
	"time"
)

func (a *API) UserRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)

	// Profile Routes
	router.Method("POST", "/change_password", Handler(a.changePassword))
	router.Method("PATCH", "/{userID}/username", Handler(a.updateUserName))
	router.Method("PATCH", "/{userID}/transaction_password", Handler(a.transactionPasswordActions))
	router.Method("PATCH", "/{userID}/update_kyc_information", Handler(a.updateKYCInformation))
	router.Method("PATCH", "/{userID}/profile", Handler(a.updateProfile))

	// Transaction Routes
	router.Method("POST", "/{userID}/transaction", Handler(a.createTransaction))
	router.Method("PATCH", "/transaction/{transactionID}/", Handler(a.updateTransaction))
	router.Method("GET", "/{userID}/transaction", Handler(a.getTransaction))
	router.Method("GET", "/transaction/{transactionID}/get_agent", Handler(a.getAgentForTransaction))

	// OTP Token Routes
	router.Method("GET", "/{userID}/otp", Handler(a.generateOTPToken))
	router.Method("POST", "/{userID}/otp", Handler(a.validateOTPToken))

	// Wallet Routes
	router.Method("POST", "/{userID}/wallet", Handler(a.createWallet))
	router.Method("PATCH", "/{userID}/wallet", Handler(a.updateWallet))
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
	authorization := strings.Split(r.Header.Get("Authorization"), " ")[1]
	if authorization == "" {
		return RespondWithError(nil, "Unauthorized", http.StatusUnauthorized, &tracingContext)
	}
	password.AccessToken = authorization

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

func (a *API) transactionPasswordActions(w http.ResponseWriter, r *http.Request) *ServerResponse {
	userID := chi.URLParam(r, "userID")
	action := r.URL.Query().Get("action")
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	user, err := a.Deps.DAL.UserDAL.FindByID(context.TODO(), userID)
	if err != nil {
		return RespondWithError(err, "could not find user", http.StatusBadRequest, &tracingContext)
	}

	switch action {
	case "create":
		if user.TransactionPassword != "" {
			return RespondWithError(nil, "user already created a transaction pin", http.StatusForbidden, &tracingContext)
		}
		var utemp model.User
		if err := decodeJSONBody(&tracingContext, r.Body, &utemp); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		if utemp.TransactionPassword == "" {
			return RespondWithError(nil, "Transaction password is required", http.StatusBadRequest, &tracingContext)
		}
		temp := &model.User{
			TransactionPassword: utemp.TransactionPassword,
		}
		password, err := helpers.HashPassword(temp.TransactionPassword)
		if err != nil {
			return RespondWithError(err, "Failed to hash password", http.StatusInternalServerError, &tracingContext)
		}
		err = a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userID, bson.D{{"$set", bson.D{{"transaction_password", password}, {"has_transaction_password", true}}}})
		if err != nil {
			return RespondWithError(err, "Failed to update transaction password", http.StatusInternalServerError, &tracingContext)
		}

		response := map[string]interface{}{
			"message": "transaction password created successfully",
		}

		return &ServerResponse{Payload: response}

	case "update":
		var request model.ChangeTransactionPasswordRequest
		if err := decodeJSONBody(&tracingContext, r.Body, &tracingContext); err != nil {
			return RespondWithError(err, "failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}
		if request.OTP == "" {
			return RespondWithError(nil, "otp is required to change transaction password", http.StatusBadRequest, &tracingContext)
		}
		if request.ProposedPassword == "" {
			return RespondWithError(nil, "new password is required", http.StatusBadRequest, &tracingContext)
		}

		valid := helpers.ValidateOTPCode(userID, request.OTP)
		if !valid {
			return RespondWithError(nil, "Invalid Token", http.StatusBadRequest, &tracingContext)
		}
		hashPassword, err := helpers.HashPassword(request.ProposedPassword)
		if err != nil {
			return RespondWithError(err, "unable to hash password. Please try again", http.StatusInternalServerError, &tracingContext)
		}

		// update utemp transaction password
		err = a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userID, bson.D{{"$set", bson.D{{
			"transaction_password", hashPassword,
		}}}})
		if err != nil {
			return RespondWithError(err, "unable to update utemp. Please try again", http.StatusInternalServerError, &tracingContext)
		}

		response := map[string]interface{}{
			"message": "token successfully generated",
		}

		return &ServerResponse{
			Payload: response,
		}

	default:
		return RespondWithWarning(nil, "this transaction action is not supported", http.StatusBadRequest, &tracingContext)
	}
}

func (a *API) updateUserName(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var param model.UpdateUsername
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	userID := chi.URLParam(r, "userID")

	if err := decodeJSONBody(&tracingContext, r.Body, &param); err != nil {
		fmt.Println(err)
		return RespondWithError(err, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	param.AccessToken = strings.Split(r.Header.Get("Authorization"), " ")[1]
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

	err = a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userID, bson.D{{"$set", bson.D{{"username", param.PreferredUsername}}}})
	if err != nil {
		return RespondWithError(err, "Failed to update user name", http.StatusInternalServerError, &tracingContext)

	}

	response := map[string]interface{}{
		"message": "transaction username updated successfully",
	}

	return &ServerResponse{Payload: response}
}

func (a *API) updateKYCInformation(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var user model.UpdateKYCInfo
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	userID := chi.URLParam(r, "userID")

	if err := decodeJSONBody(&tracingContext, r.Body, &user); err != nil {
		return RespondWithError(err, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	if user.Location == "" {
		return RespondWithError(nil, "location is required", http.StatusBadRequest, &tracingContext)
	}
	if user.Nationality == "" {
		return RespondWithError(nil, "nationality is required", http.StatusBadRequest, &tracingContext)
	}
	if user.DateOfBirth == "" {
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
	if user.IDExpiryDate == "" {
		return RespondWithError(nil, "id_expiry_date is required", http.StatusBadRequest, &tracingContext)
	}
	if user.IDImage == "" {
		return RespondWithError(nil, "id_image is required", http.StatusBadRequest, &tracingContext)
	}

	doc, err := helpers.MarshalStructToBSONDoc(user)
	if err != nil {
		return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
	}

	err = a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userID, bson.D{{"$set", doc}})
	if err != nil {
		return RespondWithError(err, "Failed to update KYC information", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"message": "KYC information successfully updated",
	}

	return &ServerResponse{Payload: response}
}

func (a *API) updateProfile(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var user model.User
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	userID := chi.URLParam(r, "userID")

	if err := decodeJSONBody(&tracingContext, r.Body, &user); err != nil {
		return RespondWithError(err, "failed to decode request body", http.StatusBadRequest, &tracingContext)
	}
	temp := &model.UpdateUserInfo{
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}
	err := a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userID, bson.D{{"$set", bson.D{{"phone_number", temp.PhoneNumber}}}})
	if err != nil {
		return RespondWithError(err, "failed to update profile", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"message": "user profile successfully updated",
	}

	return &ServerResponse{
		Payload: response,
	}
}

// Transaction

func (a *API) createTransaction(w http.ResponseWriter, r *http.Request) *ServerResponse {
	ctx := context.Background()
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	transactionType := r.URL.Query().Get("transaction-type")
	userId := chi.URLParam(r, "userID")
	user, err := a.Deps.DAL.UserDAL.FindByID(context.TODO(), userId)
	if err != nil {
		return RespondWithError(err, "Unable to get user information", http.StatusInternalServerError, &tracingContext)
	}

	// start a transaction session
	ses, err := a.Deps.DAL.Client.StartSession()
	if err != nil {
		logrus.Fatalf("[Mongo]: unable to create a session: %s", err.Error())
		return RespondWithError(nil, "Something went wrong. Please Try again", http.StatusInternalServerError, &tracingContext)
	}
	defer ses.EndSession(ctx)
	switch transactionType {
	case types.TRANSFER:
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

		err := a.Deps.DAL.TransactionDAL.CreateTransfer(context.TODO(), &transfer)
		if err != nil {
			return RespondWithError(err, "Failed to initiate transfer. Please try again", http.StatusInternalServerError, &tracingContext)
		}
		response := map[string]interface{}{
			"message": "successfully initiated transfer",
		}
		return &ServerResponse{
			Payload: response,
		}

	case types.ONE_PURSE_TRANSACTION:
		var transaction model.OnePurseTransaction
		if err := decodeJSONBody(&tracingContext, r.Body, &transaction); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		if transaction.FromUser == nil || transaction.ToUser == nil {
			return RespondWithError(nil, "sender and receiver is required", http.StatusBadRequest, &tracingContext)
		}
		if transaction.Currency == "" || transaction.Amount == 0 {
			return RespondWithError(nil, "transaction amount and currency is required", http.StatusBadRequest, &tracingContext)
		}
		if transaction.Type == "" {
			return RespondWithError(nil, "transaction type must be specified", http.StatusBadRequest, &tracingContext)
		}

		transaction.CreatedAt = time.Now()
		transaction.Status = "created"
		transaction.ID = cuid.New()

		if transaction.Type == types.REQUEST {
			_, err := ses.WithTransaction(ctx, func(sesCtx mongo.SessionContext) (interface{}, error) {
				// Create transaction
				err = a.Deps.DAL.TransactionDAL.CreateOnePurseTransaction(sesCtx, &transaction)
				if err != nil {
					return nil, errors.Wrap(err, "unable to create transaction")
				}

				// Create a Notification
				message := fmt.Sprintf("%s requested for %s %v from you", transaction.FromUser.UserName, transaction.Currency, transaction.Amount)
				err := a.CreateNotification(sesCtx, transaction.ToUser.ID, types.PAYMENT_REQUEST, message, types.ONE_PURSE_TRANSACTION, transaction.ToUser.DeviceToken, transaction)
				if err != nil {
					return nil, errors.Wrap(err, "unable to send notification")
				}

				return nil, nil
			})
			if err != nil {
				return RespondWithError(err, "something went wrong. Please try again", http.StatusInternalServerError, &tracingContext)
			}

			response := map[string]interface{}{
				"message": "successfully sent payment request",
			}
			return &ServerResponse{
				Payload: response,
			}
		} else if transaction.Type == types.PAY {
			// check that sender has enough in wallet
			pass := helpers.DoSufficientFundsCheck(user, transaction.Amount, transaction.Currency)
			if !pass {
				return RespondWithError(nil, "Insufficient funds. Please top-up wallet", http.StatusBadRequest, &tracingContext)
			}
			_, err := ses.WithTransaction(ctx, func(sesCtx mongo.SessionContext) (interface{}, error) {
				// check that user has a wallet for currency being sent, and create if not
				pass := helpers.DoUserWalletCheck(transaction.ToUser, transaction.Currency)
				if !pass {
					// create wallet for user
					wallet := &model.Wallet{
						Currency:         transaction.Currency,
						AvailableBalance: 0,
						PendingBalance:   0,
						TotalVolume:      0,
						CreatedAt:        time.Now(),
					}
					err := a.Deps.DAL.UserDAL.UpdateUser(sesCtx, transaction.ToUser.ID, bson.D{{fmt.Sprintf("wallet.%s", transaction.Currency), wallet}})
					if err != nil {
						return nil, errors.New("could not create wallet for user")
					}
				}

				// create Notification
				message := fmt.Sprintf("%s just sent %s %v to you", transaction.FromUser.UserName, transaction.Currency, transaction.Amount)
				err := a.CreateNotification(sesCtx, transaction.ToUser.ID, types.PAYMENT_RECEIVED, message, types.ONE_PURSE_TRANSACTION, transaction.ToUser.DeviceToken, transaction)
				if err != nil {
					return nil, err
				}

				// remove amount from sender's wallet
				err = a.Deps.DAL.UserDAL.UpdateUser(sesCtx, transaction.FromUser.ID, bson.D{{"$inc",
					bson.D{{
						fmt.Sprintf("wallet.%s.available_balance", transaction.Currency), -transaction.Amount,
					}}}})
				if err != nil {
					return nil, errors.New("unable to update sender's wallet")
				}

				//add amount to receivers wallet
				err = a.Deps.DAL.UserDAL.UpdateUser(sesCtx, transaction.ToUser.ID, bson.D{{"$inc",
					bson.D{{
						fmt.Sprintf("wallet.%s.available_balance", transaction.Currency), transaction.Amount,
					}}}})
				if err != nil {
					return nil, errors.New("unable to update recipient's wallet")
				}

				err = a.Deps.DAL.TransactionDAL.CreateOnePurseTransaction(sesCtx, &transaction)
				if err != nil {
					return nil, err
				}

				return nil, nil
			})
			if err != nil {
				return RespondWithError(err, "Something went wrong. Please try again", http.StatusInternalServerError, &tracingContext)
			}

			response := map[string]interface{}{
				"message": "successfully made payment",
			}
			return &ServerResponse{
				Payload: response,
			}
		}
		return RespondWithError(nil, "transaction type not supported", http.StatusBadRequest, &tracingContext)

	case types.WITHDRAW:
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
		err := a.Deps.DAL.TransactionDAL.CreateWithdrawal(context.TODO(), &withdrawal)
		if err != nil {
			return RespondWithError(err, "Failed to initiate withdrawal. Please try again", http.StatusBadRequest, &tracingContext)
		}
		response := map[string]interface{}{
			"message": "successfully initiated withdrawal",
		}
		return &ServerResponse{
			Payload: response,
		}

	case types.DEPOSIT:
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

		err := a.Deps.DAL.TransactionDAL.CreateDeposit(context.TODO(), &deposit)
		if err != nil {
			return RespondWithError(err, "Failed to initiate deposit, Please try again", http.StatusInternalServerError, &tracingContext)
		}
		response := map[string]interface{}{
			"message": "successfully initiated deposit",
		}
		return &ServerResponse{
			Payload: response,
		}

	case types.EXCHANGE:
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
		err := a.Deps.DAL.TransactionDAL.CreateExchange(context.TODO(), &exchange)
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
	case types.TRANSFER:
		var transfer model.Transfer

		if err := decodeJSONBody(&tracingContext, r.Body, &transfer); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		doc, err := helpers.MarshalStructToBSONDoc(transfer)
		if err != nil {
			return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
		}

		tempTransfer, err := a.Deps.DAL.TransactionDAL.GetTransferByID(context.TODO(), transactionId)
		if err != nil {
			return RespondWithError(err, "error fetching transfer information", http.StatusBadRequest, &tracingContext)
		}
		if tempTransfer.Status == "completed" {
			// No update should be undertaken
			return RespondWithError(nil, "cannot update a completed transaction", http.StatusForbidden, &tracingContext)
		}

		transfer.UpdatedAt = time.Now()
		err = a.Deps.DAL.TransactionDAL.UpdateTransfer(context.TODO(), transactionId, doc)
		if err != nil {
			return RespondWithError(err, "unable to update transfer information", http.StatusInternalServerError, &tracingContext)
		}

		response := map[string]interface{}{
			"message": "transfer information updated successfully",
		}
		return &ServerResponse{
			Payload: response,
		}
	case types.WITHDRAW:
		var withdrawal model.Withdrawal

		if err := decodeJSONBody(&tracingContext, r.Body, &withdrawal); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		doc, err := helpers.MarshalStructToBSONDoc(withdrawal)
		if err != nil {
			return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
		}

		tempWithdrawal, err := a.Deps.DAL.TransactionDAL.GetWithdrawalByID(context.TODO(), transactionId)
		if err != nil {
			return RespondWithError(err, "error fetching withdrawal information", http.StatusBadRequest, &tracingContext)
		}
		if tempWithdrawal.Status == "completed" {
			return RespondWithError(nil, "cannot update a completed transaction", http.StatusForbidden, &tracingContext)

		}

		withdrawal.UpdatedAt = time.Now()
		err = a.Deps.DAL.TransactionDAL.UpdateWithdrawal(context.TODO(), transactionId, doc)
		if err != nil {
			return RespondWithError(err, "unable to update withdrawal information", http.StatusInternalServerError, &tracingContext)
		}

		response := map[string]interface{}{
			"message": "withdrawal information updated successfully",
		}
		return &ServerResponse{
			Payload: response,
		}
	case types.DEPOSIT:
		var deposit model.Deposit
		if err := decodeJSONBody(&tracingContext, r.Body, &deposit); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		doc, err := helpers.MarshalStructToBSONDoc(deposit)
		if err != nil {
			return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
		}

		tempDeposit, err := a.Deps.DAL.TransactionDAL.GetDepositByID(context.TODO(), transactionId)
		if err != nil {
			return RespondWithError(err, "error fetching deposit information", http.StatusBadRequest, &tracingContext)
		}
		if tempDeposit.Status == "completed" {
			return RespondWithError(nil, "cannot update a completed transaction", http.StatusForbidden, &tracingContext)
		}

		deposit.UpdatedAt = time.Now()
		err = a.Deps.DAL.TransactionDAL.UpdateDeposit(context.TODO(), transactionId, doc)
		if err != nil {
			return RespondWithError(err, "unable to update deposit information", http.StatusInternalServerError, &tracingContext)
		}

		response := map[string]interface{}{
			"message": "update information updated successfully",
		}
		return &ServerResponse{
			Payload: response,
		}
	case types.EXCHANGE:
		var exchange model.Exchange

		if err := decodeJSONBody(&tracingContext, r.Body, &exchange); err != nil {
			return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
		}

		doc, err := helpers.MarshalStructToBSONDoc(exchange)
		if err != nil {
			return RespondWithError(err, "Failed to marshal to bson document", http.StatusInternalServerError, &tracingContext)
		}
		tempExchange, err := a.Deps.DAL.TransactionDAL.GetExchangeByID(context.TODO(), transactionId)
		if err != nil {
			return RespondWithError(err, "error fetching exchange information", http.StatusForbidden, &tracingContext)
		}
		if tempExchange.Status == "completed" {
			return RespondWithError(nil, "cannot update a completed transaction", http.StatusForbidden, &tracingContext)
		}

		exchange.UpdatedAt = time.Now()
		err = a.Deps.DAL.TransactionDAL.UpdateExchange(context.TODO(), transactionId, doc)
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
		transfers, err := a.Deps.DAL.TransactionDAL.FetchTransfers(context.TODO(), query)
		if err != nil {
			return RespondWithError(err, "unable to fetch transfers", http.StatusInternalServerError, &tracingContext)
		}
		withdraws, err := a.Deps.DAL.TransactionDAL.FetchWithdrawals(context.TODO(), query)
		if err != nil {
			return RespondWithError(err, "unable to fetch withdraws", http.StatusInternalServerError, &tracingContext)
		}
		deposits, err := a.Deps.DAL.TransactionDAL.FetchDeposits(context.TODO(), query)
		if err != nil {
			return RespondWithError(err, "unable to fetch deposits", http.StatusInternalServerError, &tracingContext)
		}
		exchanges, err := a.Deps.DAL.TransactionDAL.FetchExchanges(context.TODO(), query)
		if err != nil {
			return RespondWithError(err, "unable to fetch exchanges", http.StatusInternalServerError, &tracingContext)
		}

		response[types.TRANSFER] = transfers
		response[types.WITHDRAW] = withdraws
		response[types.DEPOSIT] = deposits
		response["exchanges"] = exchanges

		return &ServerResponse{
			Payload: response,
		}
	case types.TRANSFER:
		transfers, err := a.Deps.DAL.TransactionDAL.FetchTransfers(context.TODO(), query)
		if err != nil {
			return RespondWithError(err, "unable to fetch transfers", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: transfers,
		}
	case types.WITHDRAW:
		withdraws, err := a.Deps.DAL.TransactionDAL.FetchWithdrawals(context.TODO(), query)
		if err != nil {
			return RespondWithError(err, "unable to fetch withdraws", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: withdraws,
		}
	case types.DEPOSIT:
		deposits, err := a.Deps.DAL.TransactionDAL.FetchDeposits(context.TODO(), query)
		if err != nil {
			return RespondWithError(err, "unable to fetch deposits", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: deposits,
		}
	case types.EXCHANGE:
		exchanges, err := a.Deps.DAL.TransactionDAL.FetchExchanges(context.TODO(), query)
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
	ctx := context.Background()
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	transactionId := chi.URLParam(r, "transactionID")
	transactionType := r.URL.Query().Get("transaction-type")

	// start a transaction session
	ses, err := a.Deps.DAL.Client.StartSession()
	if err != nil {
		logrus.Fatalf("[Mongo]: unable to create a session: %s", err.Error())
		return RespondWithError(nil, "Something went wrong. Please Try again", http.StatusInternalServerError, &tracingContext)
	}
	defer ses.EndSession(ctx)

	switch transactionType {
	case types.TRANSFER:
		transfer, err := a.Deps.DAL.TransactionDAL.GetTransferByID(context.TODO(), transactionId)
		if err != nil {
			return RespondWithError(err, "could not fetch exchange information", http.StatusInternalServerError, &tracingContext)
		}
		query := bson.D{{"$gte", bson.D{{
			fmt.Sprintf("wallet.available_balance"), transfer.BaseAmount}}},
			{"wallet.currency", transfer.BaseCurrency}}
		result, err := ses.WithTransaction(ctx, func(sesCtx mongo.SessionContext) (interface{}, error) {
			agent, err := a.Deps.DAL.AgentDAL.FindOne(sesCtx, query)
			if err != nil {
				return nil, err
			}
			err = a.Deps.DAL.AgentDAL.Update(sesCtx, agent.ID, bson.D{{"$inc", bson.D{
				{"wallet.available_balance", -transfer.BaseAmount},
				{"wallet.pending_balance", transfer.BaseAmount}}}})
			if err != nil {
				return nil, err
			}
			err = a.Deps.DAL.TransactionDAL.UpdateTransfer(sesCtx, transactionId, bson.D{{"agent_account", agent}})
			if err != nil {
				return nil, err
			}

			// create notification for agent
			message := fmt.Sprintf("you have been matched to %s for a %s %v transaction", transfer.User.FullName, transfer.ConvCurrency, transfer.AmountSent)
			err = a.CreateNotification(sesCtx, agent.ID, types.TRANSACTION_MATCH, message, types.TRANSFER, agent.DeviceToken, transfer)
			if err != nil {
				return nil, err
			}

			return agent, nil
		})
		if err != nil {
			return RespondWithError(err, "unable to find an agent for your transaction right now", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: result,
		}

	case types.EXCHANGE:
		exchange, err := a.Deps.DAL.TransactionDAL.GetExchangeByID(context.TODO(), transactionId)
		if err != nil {
			return RespondWithError(err, "could not fetch exchange information", http.StatusInternalServerError, &tracingContext)
		}

		if exchange.BaseCurrency == "USD" {
			query := bson.D{{"$gte", bson.D{{
				"wallet.USD.available_balance", exchange.BaseAmount}}}}

			// start transaction
			result, err := ses.WithTransaction(ctx, func(sesCtx mongo.SessionContext) (interface{}, error) {
				user, err := a.Deps.DAL.UserDAL.FindOne(sesCtx, query)
				if err != nil {
					return nil, err
				}

				err = a.Deps.DAL.UserDAL.UpdateUser(sesCtx, user.ID, bson.D{{"$inc", bson.D{{
					fmt.Sprintf("wallet.%s.available_balance", exchange.ExchangeCurrency), -exchange.ExchangeAmount,
				}, {fmt.Sprintf("wallet.%s.pending_balance", exchange.ExchangeCurrency), exchange.ExchangeAmount}}}})
				if err != nil {
					return nil, err
				}

				// create notification
				message := fmt.Sprintf("you have been matched to %s for a %s %v transaction", exchange.User.FullName, exchange.ExchangeCurrency, exchange.ExchangeAmount)
				err = a.CreateNotification(sesCtx, user.ID, types.TRANSACTION_MATCH, message, types.EXCHANGE, user.DeviceToken, exchange)
				if err != nil {
					return nil, err
				}
				return user, nil
			})
			if err != nil {
				return RespondWithError(err, "unable to find user for your transaction now", http.StatusInternalServerError, &tracingContext)
			}
			return &ServerResponse{
				Payload: result,
			}
		} else {
			query := bson.D{{"$gte", bson.D{{
				fmt.Sprintf("wallet.available_balance"), exchange.BaseAmount}}},
				{"wallet.currency", exchange.BaseCurrency}}

			//start transaction
			result, err := ses.WithTransaction(ctx, func(sesCtx mongo.SessionContext) (interface{}, error) {
				agent, err := a.Deps.DAL.AgentDAL.FindOne(context.TODO(), query)
				if err != nil {
					return nil, err
				}

				err = a.Deps.DAL.UserDAL.UpdateUser(sesCtx, agent.ID, bson.D{{"$inc", bson.D{
					{fmt.Sprintf("wallet.%s.available_balance", exchange.ExchangeCurrency), -exchange.ExchangeAmount},
					{fmt.Sprintf("wallet.%s.pending_balance", exchange.ExchangeCurrency), exchange.ExchangeAmount},
				}}})
				if err != nil {
					return nil, err
				}

				// create notification
				message := fmt.Sprintf("you have been matched to %s for a %s %v transaction", exchange.User.FullName, exchange.ExchangeCurrency, exchange.ExchangeAmount)
				err = a.CreateNotification(sesCtx, agent.ID, types.TRANSACTION_MATCH, message, types.EXCHANGE, agent.DeviceToken, exchange)
				if err != nil {
					return nil, err
				}

				return agent, nil
			})

			if err != nil {
				return RespondWithError(err, "unable to find an agent for your transaction right now", http.StatusInternalServerError, &tracingContext)
			}
			return &ServerResponse{
				Payload: result,
			}
		}

	case types.DEPOSIT:
		deposit, err := a.Deps.DAL.TransactionDAL.GetDepositByID(context.TODO(), transactionId)
		if err != nil {
			return RespondWithError(err, "could not fetch deposit information", http.StatusInternalServerError, &tracingContext)
		}
		query := bson.D{{"$gte", bson.D{{
			fmt.Sprintf("wallet.available_balance"), deposit.BaseAmount}}},
			{"wallet.currency", deposit.BaseCurrency}}
		agent, err := a.Deps.DAL.AgentDAL.FindOne(context.TODO(), query)
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

// OTP Token Routes

func (a *API) generateOTPToken(w http.ResponseWriter, r *http.Request) *ServerResponse {
	userID := chi.URLParam(r, "userID")
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	user, err := a.Deps.DAL.UserDAL.FindByID(context.TODO(), userID)
	if err != nil {
		return RespondWithError(err, "could not find user", http.StatusBadRequest, &tracingContext)
	}

	token, err := helpers.CreateOTPCode(userID)
	if err != nil {
		return RespondWithError(err, "unable to generate otp token", http.StatusInternalServerError, &tracingContext)
	}
	err = a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userID, bson.D{{"$set", bson.D{{"otp_token", token}}}})
	if err != nil {
		return RespondWithError(err, "unable to update user token", http.StatusInternalServerError, &tracingContext)
	}

	// Send OTP token to user
	if user.PhoneNumber == "" {
		return RespondWithError(nil, "Please update phone number on profile screen", http.StatusBadRequest, &tracingContext)
	}
	msg := fmt.Sprintf("Here is your OTP for changing your transaction password: %s. It expires in %s", token, time.Now().Add(30*time.Second))
	err = a.Deps.TWILIO.SendMessage(user.PhoneNumber, msg)
	if err != nil {
		return RespondWithError(err, "unable to send otp to user. Please try again", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"message": fmt.Sprintf("token sent to %s", user.PhoneNumber),
	}

	return &ServerResponse{
		Payload: response,
	}
}

func (a *API) validateOTPToken(w http.ResponseWriter, r *http.Request) *ServerResponse {
	userID := chi.URLParam(r, "userID")
	passcode := r.URL.Query().Get("passcode")
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	_, err := a.Deps.DAL.UserDAL.FindByID(context.TODO(), userID)
	if err != nil {
		return RespondWithError(err, "could not find user", http.StatusBadRequest, &tracingContext)
	}

	valid := helpers.ValidateOTPCode(userID, passcode)
	if !valid {
		return RespondWithError(nil, "invalid otp", http.StatusBadRequest, &tracingContext)
	}

	response := map[string]interface{}{
		"message": "token validated successfully",
	}
	return &ServerResponse{
		Payload: response,
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

	user, err := a.Deps.DAL.UserDAL.FindByID(context.TODO(), userId)
	if err != nil {
		return RespondWithError(err, "unable to fetch user", http.StatusInternalServerError, &tracingContext)
	}

	wallet := model.Wallet{
		Currency:         walletType,
		AvailableBalance: 0,
		PendingBalance:   0,
		TotalVolume:      0,
		IsActive:         true,
		CreatedAt:        time.Now(),
	}

	if len(user.Wallet) == 0 {
		data := map[string]model.Wallet{
			walletType: wallet,
		}
		doc, err := helpers.MarshalStructToBSONDoc(data)
		if err != nil {
			return RespondWithError(err, "unable to marshall to mongo document", http.StatusInternalServerError, &tracingContext)
		}

		err = a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userId, bson.D{{"$set", bson.D{{
			"wallet", doc}}}})
	} else {
		doc, err := helpers.MarshalStructToBSONDoc(wallet)
		if err != nil {
			return RespondWithError(err, "unable to marshall to mongo document", http.StatusInternalServerError, &tracingContext)
		}

		err = a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userId, bson.D{{"$set", bson.D{{
			fmt.Sprintf("wallet.%s", walletType), doc}}}})
	}

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

func (a *API) updateWallet(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	walletType := r.URL.Query().Get("wallet-type")
	action := r.URL.Query().Get("action")
	userId := chi.URLParam(r, "userID")

	if walletType == "" {
		return RespondWithError(nil, "wallet type is required", http.StatusBadRequest, &tracingContext)
	}

	switch action {
	case types.DEACTIVATE:
		err := a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userId, bson.D{{"$set", bson.D{{
			fmt.Sprintf("wallet.%s.is_active", walletType), false}}}})
		if err != nil {
			return RespondWithError(err, "unable to deactivate wallet", http.StatusInternalServerError, &tracingContext)
		}
	case types.ACTIVATE:
		err := a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), userId, bson.D{{"$set", bson.D{{
			fmt.Sprintf("wallet.%s.is_active", walletType), true}}}})
		if err != nil {
			return RespondWithError(err, "unable to deactivate wallet", http.StatusInternalServerError, &tracingContext)
		}
	default:
		return RespondWithError(nil, "this action does not exist", http.StatusBadRequest, &tracingContext)
	}

	response := map[string]interface{}{
		"message": fmt.Sprintf("successfully updated %s wallet", walletType),
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
	transfers, err := a.Deps.DAL.TransactionDAL.FetchTransfers(context.TODO(), query)
	if err != nil {
		return RespondWithError(err, "unable to fetch transfers", http.StatusInternalServerError, &tracingContext)
	}
	withdraws, err := a.Deps.DAL.TransactionDAL.FetchWithdrawals(context.TODO(), query)
	if err != nil {
		return RespondWithError(err, "unable to fetch withdraws", http.StatusInternalServerError, &tracingContext)
	}
	deposits, err := a.Deps.DAL.TransactionDAL.FetchDeposits(context.TODO(), query)
	if err != nil {
		return RespondWithError(err, "unable to fetch deposits", http.StatusInternalServerError, &tracingContext)
	}
	exchanges, err := a.Deps.DAL.TransactionDAL.FetchExchanges(context.TODO(), query)
	if err != nil {
		return RespondWithError(err, "unable to fetch exchanges", http.StatusInternalServerError, &tracingContext)
	}

	response[types.TRANSFER] = transfers
	response[types.WITHDRAW] = withdraws
	response[types.DEPOSIT] = deposits
	response["exchanges"] = exchanges

	return &ServerResponse{
		Payload: response,
	}

}
