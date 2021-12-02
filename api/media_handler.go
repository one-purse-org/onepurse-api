package api

import (
	"errors"
	"github.com/aws/smithy-go"
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"net/http"
)

func (a *API) MediaRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)
	router.Method("POST", "/upload_media", Handler(a.uploadMedia))
	return router
}

func (a *API) uploadMedia(w http.ResponseWriter, r *http.Request) *ServerResponse {
	tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)

	file, header, err := r.FormFile("media")
	if err != nil {
		return RespondWithError(err, "Could not parse file", http.StatusBadRequest, &tracingContext)
	}
	location, s3err := a.Deps.AWS.S3.Upload(header.Filename, file)
	if s3err != nil {
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

	response := map[string]interface{}{
		"location": location,
	}

	return &ServerResponse{Payload: response}
}
