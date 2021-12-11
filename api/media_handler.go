package api

import (
	"errors"
	"fmt"
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
		fmt.Println(s3err)
		var ae smithy.APIError
		if errors.As(s3err, &ae) {
			fmt.Println(ae.ErrorCode())
			switch ae.ErrorCode() {
			case "AccessDenied":
				return RespondWithError(err, "Access denied", http.StatusForbidden, &tracingContext)
			case "AccountProblem":
				return RespondWithError(err, "There was a problem with the AWS account", http.StatusForbidden, &tracingContext)
			case "AllAccessDisabled":
				return RespondWithError(err, "Access to this resource has been disabled", http.StatusForbidden, &tracingContext)
			case "EntityTooSmall":
				return RespondWithError(err, "Proposed upload is smaller that minimum allowed", http.StatusBadRequest, &tracingContext)
			case "IncompleteBody":
				return RespondWithError(err, "Number of byte specified by Content-Length HTTP header not provided", http.StatusBadRequest, &tracingContext)
			default:
				return RespondWithError(err, "Could not complete file upload", http.StatusInternalServerError, &tracingContext)
			}
		}
	}

	response := map[string]interface{}{
		"location": location,
	}

	return &ServerResponse{Payload: response}
}
