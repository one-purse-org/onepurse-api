package api

import (
	"fmt"
	"github.com/aws/smithy-go"
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"github.com/lucsky/cuid"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

func (a *API) AuthRoutes(router *chi.Mux) http.Handler {
	router.Method("POST", "/login", Handler(a.login))
	router.Method("POST", "/signup", Handler(a.signUp))
	router.Method("POST", "/confirm_signup", Handler(a.confirmSignUp))

	return router
}

func (a *API) login(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var login model.LoginRequest
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &login); err != nil {
		return RespondWithError(err, "Failed to decode request body", http.StatusBadRequest, &tracingContext)
	}
	if login.Email == "" {
		return RespondWithError(nil, "Email is a required field", http.StatusBadRequest, &tracingContext)
	}

	if login.Password == "" {
		return RespondWithError(nil, "Password is a required field", http.StatusBadRequest, &tracingContext)
	}

	authResponse, err := a.Deps.AWS.Cognito.Login(&login)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			switch ae.ErrorCode() {
			case "InvalidParameterException":
				return RespondWithError(err, "Invalid parameters provided", http.StatusBadRequest, &tracingContext)
			case "NotAuthorizedException":
				return RespondWithError(err, "Not authorized", http.StatusUnauthorized, &tracingContext)
			case "PasswordResetRequiredException":
				return RespondWithError(err, "Password reset required", http.StatusUnauthorized, &tracingContext)
			case "UserNotConfirmedException":
				return RespondWithError(err, "User is not confirmed", http.StatusUnauthorized, &tracingContext)
			case "UserNotFoundException":
				return RespondWithError(err, "User is not found", http.StatusNotFound, &tracingContext)
			case "InvalidPasswordException":
				return RespondWithError(err, "Invalid password provided", http.StatusBadRequest, &tracingContext)
			default:
				return RespondWithError(err, "Could not complete request", http.StatusInternalServerError, &tracingContext)
			}
		}
	}

	return &ServerResponse{Payload: authResponse}
}

func (a *API) signUp(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var registration model.RegistrationRequest
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &registration); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	if registration.Email == "" {
		return RespondWithError(nil, "Email is a required field", http.StatusBadRequest, &tracingContext)
	}

	if registration.FullName == "" {
		return RespondWithError(nil, "first name is a required field", http.StatusBadRequest, &tracingContext)
	}

	if registration.Phone == "" {
		return RespondWithError(nil, "Phone is a required field", http.StatusBadRequest, &tracingContext)
	}

	if registration.Password == "" {
		return RespondWithError(nil, "Password is a required field", http.StatusBadRequest, &tracingContext)
	}

	signupResponse, err := a.Deps.AWS.Cognito.SignUp(&registration)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			switch ae.ErrorCode() {
			case "InvalidParameterException":
				return RespondWithError(err, "Invalid parameters provided", http.StatusBadRequest, &tracingContext)
			case "InvalidPasswordException":
				return RespondWithError(err, "Password should be at lease eight characters long, contain uppercase, lowercase characters and symbols", http.StatusBadRequest, &tracingContext)
			case "UsernameExistsException":
				return RespondWithError(err, "User already exists. Please sign in", http.StatusBadRequest, &tracingContext)
			case "CodeDeliveryFailureException":
				return RespondWithError(err, "Could not send verification code", http.StatusBadRequest, &tracingContext)
			case "NotAuthorizedException":
				return RespondWithError(err, "Not authorized", http.StatusUnauthorized, &tracingContext)
			default:
				return RespondWithError(err, fmt.Sprintf("Failed to complete signup for user : %v", registration.Email), http.StatusInternalServerError, &tracingContext)

			}
		}
	}

	user := &model.User{
		ID:        cuid.New(),
		FullName:  registration.FullName,
		Email:     registration.Email,
		CreatedAt: time.Now(),
		Active:    true,
	}

	err = a.Deps.DAL.UserDAL.Add(user)
	if err != nil {
		return RespondWithError(err, "Failed to create user", http.StatusInternalServerError, &tracingContext)
	}

	return &ServerResponse{Payload: signupResponse}
}

func (a *API) confirmSignUp(w http.ResponseWriter, r *http.Request) *ServerResponse {
	var verification model.VerificationRequest
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	if err := decodeJSONBody(&tracingContext, r.Body, &verification); err != nil {
		return RespondWithError(nil, "Failed to decode request body", http.StatusInternalServerError, &tracingContext)
	}

	if verification.Email == "" {
		return RespondWithError(nil, "Email is a required field", http.StatusBadRequest, &tracingContext)
	}

	if verification.Code == "" {
		return RespondWithError(nil, "Code is a required field", http.StatusBadRequest, &tracingContext)
	}

	status, err := a.Deps.AWS.Cognito.ConfirmSignUp(&verification)
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
		"confirmed": status,
	}

	return &ServerResponse{Payload: response}
}
