package patient

import (
	"fmt"
	"net/http"
)

// handleGetResults handles the "GET /results" endpoint and returns the results
// of an existing customer.
func (s *Server) handleGetResults(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.badRequest(res, req, "you not authorized to access this resource")
		return
	}

	results, err := s.admindb.Results(authID)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("admindb.Results error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, results)
}
