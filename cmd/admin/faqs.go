package admin

import (
	"fmt"
	"net/http"

	"github.com/ukane-philemon/labtracka-api/db"
	"github.com/ukane-philemon/labtracka-api/internal/request"
)

// handleGetFaqs handles the "GET /faqs" endpoint and returns frequently asked
// questions from the db.
func (s *Server) handleGetFaqs(res http.ResponseWriter, req *http.Request) {
	faqs, err := s.db.Faqs()
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.Faqs error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, faqs)
}

// handleUpdateFaqs handles the "POST /faqs" endpoint and updates the frequently
// asked questions from the db.
func (s *Server) handleUpdateFaqs(res http.ResponseWriter, req *http.Request) {
	var reqBody *db.Faqs
	err := request.DecodeJSONStrict(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	faqs, err := s.db.UpdateFaqs(reqBody)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.UpdateFaqs error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, faqs)
}
