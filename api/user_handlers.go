package api

import (
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

func (a *API) UserRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)
	router.Method("PATCH", "/create_transaction_password", Handler(a.createTransactionPassword))

	return router
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
