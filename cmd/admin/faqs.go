package admin

import (
	"fmt"
	"net/http"
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
