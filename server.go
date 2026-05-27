package main

import (
	"go-backend/transcribe"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	handler http.Handler
}

func NewServer(userController *transcribe.Controller) *Server {
	mux := http.NewServeMux()
	userController.RegisterRoutes(mux)
	mux.Handle("GET /metrics", promhttp.Handler())
	return &Server{
		handler: otelhttp.NewHandler(mux, "http.server"),
	}
}

func (s *Server) Run(addr string) error {
	slog.Info("Server started", "addr", addr)
	return http.ListenAndServe(addr, s.handler)
}
