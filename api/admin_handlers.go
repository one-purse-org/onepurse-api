package api

import (
	"context"
	"fmt"
	"github.com/aws/smithy-go"
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"github.com/isongjosiah/work/onepurse-api/types"
	"github.com/lucsky/cuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
)

func (a *API) AdminRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)
	router.Method("POST", "/create_admin", Handler(a.createAdmin))
	router.Method("POST", "/create_currency", Handler(a.createCurrency))

	/*Dashboard*/
	router.Method("GET", "/transaction/get_metrics", Handler(a.getMetrics))
	router.Method("GET", "/transactions/get_volume", Handler(a.getTransactionVolume))

	/*User*/
	router.Method("GET", "/user", Handler(a.getAllUsers))
	router.Method("PATCH", "/user/action", Handler(a.userActions))
	router.Method("GET", "/user/transaction_history", Handler(a.getUserTransactionHistory))

	/*AGENT*/
	router.Method("POST", "/agent", Handler(a.adminCreateAgent))
	router.Method("GET", "/agent", Handler(a.getAllAgents))
	router.Method("PATCH", "/agent/action", Handler(a.agentActions))
	router.Method("GET", "/agent/transaction_history", Handler(a.getAgentTransactionHistory))

	/*TRANSACTION*/
	router.Method("GET", "/transaction", Handler(a.fetchAllTransactions))

	/*EXCHANGE RATE*/
	router.Method("PATCH", "/exchange_rate", Handler(a.updateExchangeRate))

	/*ADMIN PAYMENT*/
	router.Method("POST", "/payments", Handler(a.createAdminPayments))
	router.Method("GET", "/payments", Handler(a.fetchAdminPayments))

	/*SETTINGS*/
	router.Method("GET", "/{ID}/security/generate_code", Handler(a.generateSecurityCode))
	router.Method("PATCH", "/security/update_profile", Handler(a.updateAdminProfile))

	return router
}

// createAdmin ...
func (a *API) createAdmin(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var admin model.Admin
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &admin); err != nil {
		fmt.Println(err)
		return RespondWithError(err, "Failed to decode request body", http.StatusBadRequest, &tracingContext)
	}

	if admin.FullName == "" {
		return RespondWithError(nil, "full_name is required", http.StatusBadRequest, &tracingContext)
	}
	if admin.Username == "" {
		return RespondWithError(nil, "username is required", http.StatusBadRequest, &tracingContext)
	}
	if admin.Email == "" {
		return RespondWithError(nil, "email is required", http.StatusBadRequest, &tracingContext)
	}
	if admin.Phone == "" {
		return RespondWithError(nil, "phone is required", http.StatusBadRequest, &tracingContext)
	}
	if admin.Role == nil {
		return RespondWithError(nil, "admin role is required", http.StatusBadRequest, &tracingContext)
	}

	registration := model.CreateUserRequest{
		Email:    admin.Email,
		FullName: admin.FullName,
		Phone:    admin.Phone,
		UserName: admin.Username,
	}
	createResponse, err := a.Deps.AWS.Cognito.CreateUser(&registration)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) { //TODO(JOSIAH): Verify the errors thrown
			switch ae.ErrorCode() {
			case "InvalidParameterException":
				return RespondWithError(err, "Invalid parameters provided", http.StatusBadRequest, &tracingContext)
			case "InvalidPasswordException":
				return RespondWithError(err, "Password should be at lease eight characters long, contain uppercase, lowercase characters and symbols", http.StatusBadRequest, &tracingContext)
			case "UsernameExistsException":
				return RespondWithError(err, "UserID already exists. Please sign in", http.StatusBadRequest, &tracingContext)
			case "CodeDeliveryFailureException":
				return RespondWithError(err, "Could not send verification code", http.StatusBadRequest, &tracingContext)
			case "NotAuthorizedException":
				return RespondWithError(err, "Not authorized", http.StatusUnauthorized, &tracingContext)
			default:
				return RespondWithError(err, fmt.Sprintf("Failed to complete signup for user : %v", registration.Email), http.StatusInternalServerError, &tracingContext)

			}
		}
	}

	admin.ID = cuid.New()
	err = a.Deps.DAL.AdminDAL.AddAdmin(context.TODO(), &admin)
	if err != nil {
		return RespondWithError(err, "Failed to create admin", http.StatusInternalServerError, &tracingContext)
	}

	return &ServerResponse{
		Payload:    createResponse,
		Message:    "admin created successfully",
		StatusCode: http.StatusCreated,
	}
}

// createCurrency allows an authorized admin to create currency available on the platform
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

// getMetrics allows an authorized admin fetch defined metrics on the platform
func (a *API) getMetrics(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	// Fetch NumberMetrics
	numberMetric, err := a.GetNumberMetrics(context.TODO())
	if err != nil {
		return RespondWithError(err, "unable to fetch number metrics", http.StatusInternalServerError, &tracingContext)
	}

	// Fetch TransactionVolumeMetrics
	transactionMetric, err := a.TransactionVolumeMetrics(context.TODO(), time.Time{}, time.Time{})
	if err != nil {
		return RespondWithError(err, "unable to fetch transaction metrics", http.StatusInternalServerError, &tracingContext)
	}

	// Fetch CurrencyMetrics
	currencyMetrics, err := a.GetCurrencyMetrics(context.TODO())
	if err != nil {
		return RespondWithError(err, "unable to fetch currency metrics", http.StatusInternalServerError, &tracingContext)
	}

	metrics := model.Metric{
		NumberMetric:      numberMetric,
		TransactionMetric: transactionMetric,
		CurrencyMetric:    currencyMetrics,
	}
	return &ServerResponse{
		Payload:    metrics,
		Message:    "Metrics Fetched Successfully",
		StatusCode: http.StatusOK,
	}
}

// getTransactionVolume allows an authorized admin fetch the transaction volume by a defined time range
func (a *API) getTransactionVolume(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	var start time.Time
	var end time.Time
	var err error
	if startDate != "" {
		start, err = time.Parse("2002-01-23", startDate)
		if err != nil {
			return RespondWithError(err, "unable to parse start date", http.StatusBadRequest, &tracingContext)
		}
	}
	if endDate != "" {
		end, err = time.Parse("2002-01-23", endDate)
		if err != nil {
			return RespondWithError(err, "unable to parse end date", http.StatusBadRequest, &tracingContext)
		}
	}

	metrics, err := a.TransactionVolumeMetrics(context.TODO(), start, end)
	if err != nil {
		return RespondWithError(err, "unable to fetch transaction volume metrics", http.StatusInternalServerError, &tracingContext)
	}

	return &ServerResponse{
		Payload: metrics,
	}
}

// getAllUsers allows an authorized admin fetch all the users on the platform
func (a *API) getAllUsers(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	fetchType := r.URL.Query().Get("type")

	var user *[]model.User
	var err error
	switch fetchType {
	case "all":
		user, err = a.Deps.DAL.UserDAL.FindAll(context.TODO(), bson.D{})
		if err != nil {
			return RespondWithError(err, "unable to fetch all users", http.StatusInternalServerError, &tracingContext)
		}
	case types.REJECTED:
		user, err = a.Deps.DAL.UserDAL.FindAll(context.TODO(), bson.D{{"approved", false}})
		if err != nil {
			return RespondWithError(err, "unable to fetch rejected users", http.StatusInternalServerError, &tracingContext)
		}
	case types.APPROVED:
		user, err = a.Deps.DAL.UserDAL.FindAll(context.TODO(), bson.D{{"approved", true}})
		if err != nil {
			return RespondWithError(err, "unable to fetch approved users", http.StatusInternalServerError, &tracingContext)
		}
	case types.SINGLE:
		id := r.URL.Query().Get("id")
		suser, err := a.Deps.DAL.UserDAL.FindOne(context.TODO(), bson.D{{"_id", id}})
		if err != nil {
			return RespondWithError(err, "unable to fetch single user", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Message: "User fetched successfully",
			Payload: suser,
		}

	default:
		return &ServerResponse{
			Message:    "Specified type is not supported",
			StatusCode: http.StatusBadRequest,
			Err:        errors.New("Specified type is not supported"),
		}
	}
	return &ServerResponse{
		Payload:    user,
		Message:    "users fetched successfully",
		StatusCode: http.StatusOK,
	}
}

// userActions allows an authorized admin approve or reject a user
func (a *API) userActions(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	action := r.URL.Query().Get("action")
	id := r.URL.Query().Get("id")

	switch action {
	case types.APPROVE:
		err := a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), id, bson.D{{"$set", bson.D{{"approved", true}}}})
		if err != nil {
			return RespondWithError(err, "unable to approve user", http.StatusInternalServerError, &tracingContext)
		}
	case types.REJECT:
		err := a.Deps.DAL.UserDAL.UpdateUser(context.TODO(), id, bson.D{{"$set", bson.D{{"approved", false}}}})
		if err != nil {
			return RespondWithError(err, "unable to reject user", http.StatusInternalServerError, &tracingContext)
		}
	default:
		return &ServerResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "action type is not specified",
			Err:        errors.New("action type is not specified"),
		}
	}

	return &ServerResponse{
		Message: "user updated successfully",
	}
}

// getUserTransactionHistory allows an authorized admin fetch a user transaction history
func (a *API) getUserTransactionHistory(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	id := r.URL.Query().Get("id")
	query := bson.D{{"user_id", id}}

	response := make(map[string]interface{})
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

	if len(*transfers) == 0 {
		transfers = &[]model.Transfer{}
	}
	if len(*withdraws) == 0 {
		withdraws = &[]model.Withdrawal{}
	}
	if len(*deposits) == 0 {
		deposits = &[]model.Deposit{}
	}
	if len(*exchanges) == 0 {
		exchanges = &[]model.Exchange{}
	}

	response[types.TRANSFER] = transfers
	response[types.WITHDRAW] = withdraws
	response[types.DEPOSIT] = *deposits
	response[types.EXCHANGE] = *exchanges

	return &ServerResponse{
		Payload: response,
		Message: "user transaction successfully",
	}
}

//adminCreateAgent allows an authorized admin create an agent
func (a *API) adminCreateAgent(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var agent model.Agent
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &agent); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusBadRequest, &tracingContext)
	}

	if agent.FullName == "" {
		return RespondWithError(nil, "full_name is required", http.StatusBadRequest, &tracingContext)
	}
	if agent.Email == "" {
		return RespondWithError(nil, "email is required", http.StatusBadRequest, &tracingContext)
	}
	if agent.Phone == "" {
		return RespondWithError(nil, "phone is required", http.StatusBadRequest, &tracingContext)
	}

	registration := model.CreateUserRequest{
		Email:    agent.Email,
		FullName: agent.FullName,
		Phone:    agent.Phone,
		UserName: agent.UserName,
	}
	createResponse, err := a.Deps.AWS.Cognito.CreateUser(&registration)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) { //TODO(JOSIAH): Verify the errors thrown
			switch ae.ErrorCode() {
			case "InvalidParameterException":
				return RespondWithError(err, "Invalid parameters provided", http.StatusBadRequest, &tracingContext)
			case "InvalidPasswordException":
				return RespondWithError(err, "Password should be at lease eight characters long, contain uppercase, lowercase characters and symbols", http.StatusBadRequest, &tracingContext)
			case "UsernameExistsException":
				return RespondWithError(err, "UserID already exists. Please sign in", http.StatusBadRequest, &tracingContext)
			case "CodeDeliveryFailureException":
				return RespondWithError(err, "Could not send verification code", http.StatusBadRequest, &tracingContext)
			case "NotAuthorizedException":
				return RespondWithError(err, "Not authorized", http.StatusUnauthorized, &tracingContext)
			default:
				return RespondWithError(err, fmt.Sprintf("Failed to complete signup for user : %v", registration.Email), http.StatusInternalServerError, &tracingContext)

			}
		}
	}

	agent.ID = cuid.New()
	err = a.Deps.DAL.AgentDAL.Add(context.TODO(), &agent)
	if err != nil {
		return RespondWithError(err, "Failed to create agent", http.StatusInternalServerError, &tracingContext)
	}

	return &ServerResponse{
		Payload: createResponse,
		Message: "agent created successfully",
	}
}

// getAllAgents allows an authorized admin fetch all the agents on the platform
func (a *API) getAllAgents(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	fetchType := r.URL.Query().Get("type")

	var agent *[]model.Agent
	var err error
	switch fetchType {
	case "all":
		agent, err = a.Deps.DAL.AgentDAL.FindAll(context.TODO(), bson.D{})
		if err != nil {
			return RespondWithError(err, "unable to fetch all users", http.StatusInternalServerError, &tracingContext)
		}
	case types.REJECTED:
		agent, err = a.Deps.DAL.AgentDAL.FindAll(context.TODO(), bson.D{{"approved", false}})
		if err != nil {
			return RespondWithError(err, "unable to fetch rejected users", http.StatusInternalServerError, &tracingContext)
		}
	case types.APPROVED:
		agent, err = a.Deps.DAL.AgentDAL.FindAll(context.TODO(), bson.D{{"approved", true}})
		if err != nil {
			return RespondWithError(err, "unable to fetch approved users", http.StatusInternalServerError, &tracingContext)
		}
	case types.SINGLE:
		id := r.URL.Query().Get("id")
		sagent, err := a.Deps.DAL.AgentDAL.FindOne(context.TODO(), bson.D{{"_id", id}})
		if err != nil {
			return RespondWithError(err, "unable to fetch single agent", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Message: "agent fetched successfully",
			Payload: sagent,
		}
	default:
		return &ServerResponse{
			Err:        errors.New("this type isn't supported"),
			Message:    "type isn't supported",
			StatusCode: http.StatusBadRequest,
		}
	}
	return &ServerResponse{
		Payload: agent,
		Message: "agent fetched successfully",
	}
}

// agentActions allows an authorized admin approve or reject and agent\
func (a *API) agentActions(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	action := r.URL.Query().Get("action")
	id := r.URL.Query().Get("id")

	switch action {
	case types.APPROVE:
		err := a.Deps.DAL.AgentDAL.Update(context.TODO(), id, bson.D{{"$set", bson.D{{"approved", true}}}})
		if err != nil {
			return RespondWithError(err, "unable to approve user", http.StatusInternalServerError, &tracingContext)
		}
	case types.REJECT:
		err := a.Deps.DAL.AgentDAL.Update(context.TODO(), id, bson.D{{"$set", bson.D{{"approved", false}}}})
		if err != nil {
			return RespondWithError(err, "unable to reject user", http.StatusInternalServerError, &tracingContext)
		}
	}
	return &ServerResponse{
		Message: "agent updated successfully",
	}
}

// getAgentTransactionHistory allows an authorized admin fetch an agent transaction history
func (a *API) getAgentTransactionHistory(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	id := r.URL.Query().Get("id")
	query := bson.D{{"agent_id", id}}

	response := make(map[string]interface{})
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

	if len(*transfers) == 0 {
		transfers = &[]model.Transfer{}
	}
	if len(*withdraws) == 0 {
		withdraws = &[]model.Withdrawal{}
	}
	if len(*deposits) == 0 {
		deposits = &[]model.Deposit{}
	}
	if len(*exchanges) == 0 {
		exchanges = &[]model.Exchange{}
	}

	response[types.TRANSFER] = transfers
	response[types.WITHDRAW] = withdraws
	response[types.DEPOSIT] = deposits
	response[types.EXCHANGE] = exchanges

	return &ServerResponse{
		Payload: response,
		Message: "agent transaction fetched successfully",
	}
}

// fetchAllTransaction allows an authorized admin fetch all transactions
func (a *API) fetchAllTransactions(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	query := bson.D{}

	response := make(map[string]interface{})
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

	if len(*transfers) == 0 {
		transfers = &[]model.Transfer{}
	}
	if len(*withdraws) == 0 {
		withdraws = &[]model.Withdrawal{}
	}
	if len(*deposits) == 0 {
		deposits = &[]model.Deposit{}
	}
	if len(*exchanges) == 0 {
		exchanges = &[]model.Exchange{}
	}

	response[types.TRANSFER] = transfers
	response[types.WITHDRAW] = withdraws
	response[types.DEPOSIT] = deposits
	response[types.EXCHANGE] = exchanges

	return &ServerResponse{
		Payload: response,
		Message: "all transactions fetched successfully",
	}
}

//updateExchangeRate allows an authorized admin update the exchange rate
func (a *API) updateExchangeRate(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	var exchangeRate *model.Rate

	if err := decodeJSONBody(&tracingContext, r.Body, &exchangeRate); err != nil {
		logrus.Errorf(err.Error())
		return RespondWithError(err, "failed to decode request body", http.StatusBadRequest, &tracingContext)
	}

	return &ServerResponse{
		Payload: nil,
	}
}

// createAdminPayments allows an authorized admin create an admin payment
func (a *API) createAdminPayments(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var payment model.AdminPayment
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &payment); err != nil {
		logrus.Errorf(err.Error())
		return RespondWithError(err, "failed to decode request body", http.StatusBadRequest, &tracingContext)
	}

	err := a.Deps.DAL.TransactionDAL.CreateAdminPayment(context.TODO(), &payment)
	if err != nil {
		return RespondWithError(err, "failed to create admin payment", http.StatusInternalServerError, &tracingContext)
	}
	return &ServerResponse{
		Message: "unable to create admin payment",
	}
}

// fetchAdminPayments allows an authorized admin fetch all admin payments
func (a *API) fetchAdminPayments(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	id := r.URL.Query().Get("id")

	if id == "" {
		payments, err := a.Deps.DAL.TransactionDAL.FetchAdminPayments(context.TODO(), bson.D{})
		if err != nil {
			return RespondWithError(err, "unable to fetch admin payments", http.StatusInternalServerError, &tracingContext)
		}
		if len(*payments) == 0 {
			payments = &[]model.AdminPayment{}
		}
		return &ServerResponse{
			Payload: payments,
			Message: "admin payments fetched successfully",
		}
	} else {
		payment, err := a.Deps.DAL.TransactionDAL.GetAdminPayment(context.TODO(), bson.D{{"_id", id}})
		if err != nil {
			return RespondWithError(err, "unable to fetch admin payments", http.StatusInternalServerError, &tracingContext)
		}
		return &ServerResponse{
			Payload: payment,
			Message: "admin payment fetched successfully",
		}
	}
}

// generateSecurityCode allows an authorized admin generate a security code for sensitive actions
func (a *API) generateSecurityCode(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
	id := chi.URLParam(r, "ID")
	fmt.Println("admin", id)

	return &ServerResponse{
		Payload: tracingContext,
	}
}

// updateAdminProfile allows an authorized admin update their profile
func (a *API) updateAdminProfile(w http.ResponseWriter, r *http.Request) *ServerResponse {
	return &ServerResponse{
		Payload: nil,
	}
}
