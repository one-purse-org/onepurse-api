package api

import (
	"github.com/go-chi/chi"
	"net/http"
)

func (a *API) AgentRoutes() http.Handler {
	router := chi.NewRouter()
	router.Use(Authorization)
	router.Method("POST", "/create", Handler(a.createAgent))
	return router
}

func (a *API) createAgent(w http.ResponseWriter, r *http.Request) *ServerResponse {
	return &ServerResponse{
		Payload: nil,
	}
}
