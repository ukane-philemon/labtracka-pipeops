package patient

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ukane-philemon/labtracka-api/db"
	"github.com/ukane-philemon/labtracka-api/internal/files"
	"github.com/ukane-philemon/labtracka-api/internal/request"
	"github.com/ukane-philemon/labtracka-api/internal/validator"
)

var maxFileUploadSize = 2 * 1024 // 2MB

// handleCreateAccount handles the "POST /create-account" endpoint and creates
// an account or a new patient.
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

	// TODO: Validate deviceID

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
		Patient:  &reqBody.PatientInfo,
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

// handleGetProfile handles the "GET /profile" endpoint and returns patient
// information.
func (s *Server) handleGetProfile(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	patientInfo, err := s.db.PatientInfo(authID)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.PatientInfo error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, patientInfo)
}

func trimErrorInvalidRequest(err error) string {
	return strings.TrimPrefix(err.Error(), db.ErrorInvalidRequest.Error()+": ")
}

// handleProfileImageUpload handles the "POST /profile-image" endpoint and
// uploads or updates a patient's profile image.
func (s *Server) handleProfileImageUpload(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	err := req.ParseMultipartForm(int64(maxFileUploadSize))
	if err != nil {
		s.badRequest(res, req, "failed to parse file (max file size 2MB)")
		return
	}

	const profileImageUploadKey = "profile-image"
	f, _, err := req.FormFile(profileImageUploadKey)
	if err != nil {
		s.badRequest(res, req, "failed to parse file")
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(f)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("buf.ReadFrom error: %w", err))
		return
	}

	fileType := http.DetectContentType(buf.Bytes())
	if !files.IsSupportedImageFile(fileType) {
		s.badRequest(res, req, "only image files are supported")
		return
	}

	patientID, err := s.db.PatientID(authID)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.PatientID error: %w", err))
		return
	}

	fileURL, err := s.cfg.Uploader.UploadFile(s.ctx, patientID, "profile-image", bytes.NewReader(buf.Bytes()))
	if err != nil {
		s.serverError(res, req, fmt.Errorf("Uploader.UploadFile error: %w", err))
		return
	}

	err = s.db.SaveProfileImage(authID, fileURL)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.SaveProfileImage error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, map[string]string{
		"profile_url": fileURL,
	})
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
