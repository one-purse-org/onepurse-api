package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/isongjosiah/work/onepurse-api/common"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/isongjosiah/work/onepurse-api/deps"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type API struct {
	Server *http.Server
	Config *config.Config
	Deps   *deps.Dependencies
}

type Handler func(w http.ResponseWriter, r *http.Request) *ServerResponse

func (a *API) SetupServerHandler() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(60 * time.Second))
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"}, // TODO: use a proper origin for production app
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept", "Authorization", "Content-Type",
			"X-CSRF-Token", config.HeaderRequestID, config.HeaderRequestSource,
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Use(RequestTracing)
	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	return mux
}

func (a *API) Serve() error {
	a.Server = &http.Server{
		Addr:           fmt.Sprintf(":%d", a.Config.Port),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		Handler:        a.SetupServerHandler(),
		MaxHeaderBytes: 1024 * 1024,
	}

	logrus.Infof("[API]: running on port %v ...", a.Config.Port)

	return a.Server.ListenAndServe()
}

func (a *API) Shutdown() error {
	return a.Server.Shutdown(context.Background())
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := h(w, r)

	if response.StatusCode == 0 && response.Err == nil {
		response.StatusCode = http.StatusOK
	}

	var responseBytes []byte
	var err error
	var marshalErr error

	if response.Err != nil {
		responseBytes, marshalErr = json.Marshal(map[string]interface{}{
			"error": ErrorResponse{
				ErrorMessage: response.Message,
				ErrorCode:    response.StatusCode,
			},
		})
	} else {
		responseBytes, marshalErr = json.Marshal(response.Payload)
	}

	if marshalErr != nil {
		tracingContext := r.Context().Value(tracing.ContextKeyTracing).(tracing.Context)
		response = RespondWithError(err, "Error creating json response", http.StatusInternalServerError, &tracingContext)

		responseBytes, _ = json.Marshal(map[string]interface{}{
			"error": ErrorResponse{
				ErrorMessage: response.Message,
				ErrorCode:    response.StatusCode,
			},
		})
	}

	switch response.ContentType {
	case common.ContentTypeJSON:
		WriteJSONResponse(w, response.StatusCode, responseBytes)
	default:
		WriteJSONResponse(w, response.StatusCode, responseBytes)
	}
}

func decodeJSONBody(ctx *tracing.Context, body io.ReadCloser, target interface{}) error {
	defer body.Close()

	if body == nil {
		return fmt.Errorf("missing request body for request: %v", ctx)
	}

	if err := json.NewDecoder(body).Decode(&target); err != nil {
		return errors.Wrapf(err, "error parsing json body for request: %v", ctx)
	}

	return nil
}
