package api

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"net/http"
)

func (a *API) AdminRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)
	router.Method("POST", "/create_currency", Handler(a.createCurrency))

	return router
}

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
