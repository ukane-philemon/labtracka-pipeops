package patient

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ukane-philemon/labtracka-api/db"
	"github.com/ukane-philemon/labtracka-api/internal/request"
	"github.com/ukane-philemon/labtracka-api/internal/validator"
)

// handleCreateAccount handles the "POST /create-account" endpoint and creates
// an account or a new customer.
func (s *Server) handleCreateAccount(res http.ResponseWriter, req *http.Request) {
	var reqBody *createAccountRequest
	err := request.DecodeJSON(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	// Validate the request data.
	err = reqBody.Validate()
	if err != nil {
		s.badRequest(res, req, err.Error())
		return
	}

	// Validate password.
	if !validator.IsPasswordValid(reqBody.Password) {
		s.badRequest(res, req, validator.PassWordErrorMsg)
		return
	}

	// Check email validation token.
	if !s.optManager.ValidateOTPValidationToken(reqBody.DeviceID, reqBody.EmailValidationToken, reqBody.Email) {
		s.badRequest(res, req, "invalid OTP")
		return
	}

	// if reqBody.PublicKey == nil || reqBody.DecryptedToken == "" {
	// 	s.badRequest(res, req, "missing required encryption data")
	// 	return
	// }

	// if !s.publicKeyValidator.IsValidPublicKeyToken(reqBody.DecryptedToken, reqBody.PublicKey) {
	// 	s.badRequest(res, req, "invalid encryption data")
	// 	return
	// }

	err = s.db.CreateAccount(&db.CreateAccountRequest{
		Customer: &reqBody.CustomerInfo,
		Password: reqBody.Password,
		DeviceID: reqBody.DeviceID,
	})
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.CreateAccount error: %v", err))
		}
		return
	}

	s.optManager.DeleteOTPRecord(reqBody.DeviceID)
	// s.publicKeyValidator.DeleteToken(reqBody.DecryptedToken)

	s.sendSuccessResponse(res, req, "Account created successfully, please proceed to login")
}

func trimErrorInvalidRequest(err error) string {
	return strings.TrimPrefix(err.Error(), db.ErrorInvalidRequest.Error()+": ")
}

/*** Sub Accounts ***/

// handleAddSubAccount handles the "POST /sub-account" endpoint and adds
// a sub account to a patients account.
func (s *Server) handleAddSubAccount(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	var reqBody *db.SubAccount
	err := request.DecodeJSON(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	// Validate the request data.
	err = reqBody.Validate()
	if err != nil {
		s.badRequest(res, req, err.Error())
		return
	}

	subAccounts, err := s.db.AddSubAccount(authID, reqBody)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.AddSubAccount error: %w", err))
		}
		return
	}

	s.sendSuccessResponseWithData(res, req, subAccounts)
}

// handleGetSubAccounts handles the "GET /sub-accounts" endpoint and retrieves
// all sub accounts for a patients account.
func (s *Server) handleGetSubAccounts(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	subAccounts, err := s.db.SubAccounts(authID)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.SubAccounts error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, subAccounts)
}

// handleRemoveSubAccount handles the "DELETE /sub-account" endpoint and removes
// a sub account from a patients account. Expects an subAccountID query param.
func (s *Server) handleRemoveSubAccount(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	subAccountID := req.URL.Query().Get("subAccountID")
	if subAccountID == "" {
		s.badRequest(res, req, `subAccountID query param is required`)
		return
	}

	subAccounts, err := s.db.RemoveSubAccount(authID, subAccountID)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.RemoveSubAccount error: %w", err))
		}
		return
	}

	s.sendSuccessResponseWithData(res, req, subAccounts)
}
