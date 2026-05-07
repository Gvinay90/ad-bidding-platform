// Package router wires HTTP routes and middleware.
package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/api-gateway/handler"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// New returns a configured *http.ServeMux.
// The Swagger UI is served at /swagger/.
func New(h *handler.Handler, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /healthz", h.Health)

	// Campaigns
	mux.HandleFunc("POST /campaigns", h.CreateCampaign)
	mux.HandleFunc("GET /campaigns", h.ListCampaigns)
	mux.HandleFunc("GET /campaigns/{id}", h.GetCampaign)
	mux.HandleFunc("PUT /campaigns/{id}", h.UpdateCampaign)
	mux.HandleFunc("DELETE /campaigns/{id}", h.DeleteCampaign)

	// Bidder
	mux.HandleFunc("POST /bid", h.EvaluateBid)

	// Analytics
	mux.HandleFunc("GET /stats/{id}", h.GetCampaignStats)

	// Swagger UI  – served from /swagger/
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	return logging(log)(mux)
}

// logging is a minimal structured-logging middleware.
func logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(rw, r)
			log.Info("http",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.code,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}
