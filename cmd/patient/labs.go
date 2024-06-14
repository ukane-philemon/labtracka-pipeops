package patient

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ukane-philemon/labtracka-api/db"
)

// handleGetLabs handles the "GET /labs" endpoint and returns a list of
// supported labs.
func (s *Server) handleGetLabs(res http.ResponseWriter, req *http.Request) {
	labs, err := s.adminDB.Labs()
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.Labs error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, labs)
}

// handleGetLabTests handles the "GET /labs/{labID}/tests" returns tests and
// test packages available for lab with the url param labID.
func (s *Server) handleGetLabTests(res http.ResponseWriter, req *http.Request) {
	labID := chi.URLParam(req, "labID")
	labTests, err := s.adminDB.LabTests(labID)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.LabTests error: %w", err))
		}
		return
	}

	s.sendSuccessResponseWithData(res, req, labTests)
}
