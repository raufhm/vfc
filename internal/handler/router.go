package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func SetupRouter(handler *ProductHandler, logger *zap.Logger) *mux.Router {
	router := mux.NewRouter()

	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware)

	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/events", handler.CreateEvent).Methods("POST")
	router.HandleFunc("/products/{id}", handler.GetProduct).Methods("GET")

	return router
}

func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Info("Request started",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path))

			next.ServeHTTP(w, r)

			logger.Info("Request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)))
		})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
