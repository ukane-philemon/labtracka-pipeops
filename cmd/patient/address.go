package patient

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ukane-philemon/labtracka-api/db"
	"github.com/ukane-philemon/labtracka-api/internal/request"
)

func (s *Server) handleAddAddress(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	var reqBody *db.CustomerAddress
	err := request.DecodeJSON(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	addresses, err := s.db.AddNewAddress(authID, reqBody)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.AddNewAddress error: %w", err))
		}
		return
	}

	s.sendSuccessResponseWithData(res, req, addresses)
}
