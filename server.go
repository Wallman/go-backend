package main

import (
	"go-backend/transcribe"
	"log"
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
	log.Printf("Server started on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}
