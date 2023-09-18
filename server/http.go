package server

import (
	"encoding/json"
	"net/http"
)

type httpServer struct {
	mux *http.ServeMux
}

type erroMesage struct {
	Message string `json:"message"`
}

func NewHttpServer() *httpServer {
	return &httpServer{
		mux: http.NewServeMux(),
	}
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *httpServer) intercept(method string, next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if method != r.Method {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(erroMesage{Message: "Method not allowed"})
			return
		}
		next(w, r)
	}
}

func (s *httpServer) Post(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	handlerFunc := s.intercept(http.MethodPost, handler)
	s.mux.HandleFunc(pattern, handlerFunc)
}

func (s *httpServer) Get(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	handlerFunc := s.intercept(http.MethodGet, handler)
	s.mux.HandleFunc(pattern, handlerFunc)
}
