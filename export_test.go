package main

import "net/http"

func (s *Server) Handler() http.Handler {
	return s.mux
}
