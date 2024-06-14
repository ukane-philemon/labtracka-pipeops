package admin

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/ukane-philemon/labtracka-api/internal/response"
)

func (s *Server) registerRoutes() http.Handler {
	mux := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET", "DELETE"},
		Debug:          true,
	})
	mux.Use(cors.Handler)
	mux.Use(httprate.LimitByRealIP(20, time.Minute))

	mux.Get("/", s.root)

	return mux
}

func (s *Server) root(res http.ResponseWriter, _ *http.Request) {
	response.JSON(res, http.StatusOK, "Welcome to the Labtracka Admin API, visit https://labtracka.com to get started!")
}
