package api

import (
	"github.com/aws/smithy-go"
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/helpers"
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
	router.Method("POST", "/{userID}/create_username", Handler(a.createUserName))
	router.Method("PATCH", "/{userID}/create_transaction_password", Handler(a.createTransactionPassword))
	router.Method("PATCH", "/{userID}/update_kyc_information", Handler(a.updateKYCInformation))

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
