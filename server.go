package main

import (
	"go-backend/transcribe"
	"net/http"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer(userController *transcribe.Controller) *Server {
	s := &Server{mux: http.NewServeMux()}
	userController.RegisterRoutes(s.mux)
	return s
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}
