package api

import (
	"github.com/aws/smithy-go"
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
)

func (a *API) UserRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)
	router.Method("POST", "/change_password", Handler(a.changePassword))
	router.Method("PATCH", "/create_transaction_password", Handler(a.createTransactionPassword))

	return router
}

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
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &user); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	if user.ID == "" {
		return RespondWithError(nil, "User ID is required", http.StatusBadRequest, &tracingContext)
	}
	if user.TransactionPassword == "" {
		return RespondWithError(nil, "Transaction password is required", http.StatusBadRequest, &tracingContext)
	}

	err := a.Deps.DAL.UserDAL.UpdateUser(user.ID, bson.D{{"$set", bson.D{{"transaction_password", user.TransactionPassword}}}})
	if err != nil {
		return RespondWithError(err, "Failed to update transaction password", http.StatusInternalServerError, &tracingContext)
	}

	response := map[string]interface{}{
		"message": "transaction password updated",
	}

	return &ServerResponse{Payload: response}
}
