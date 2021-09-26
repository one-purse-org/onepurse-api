package api

import (
	"net/http"
	"nimr-backend/config"
	"nimr-backend/deps"
)

type API struct {
	Server *http.Server
	Config *config.Config
	Deps   *deps.Dependencies
}

type Handler func(w http.ResponseWriter, r *http.Request) *ServerResponse
