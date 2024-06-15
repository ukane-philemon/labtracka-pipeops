package admin

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ukane-philemon/labtracka-api/db"
	"github.com/ukane-philemon/labtracka-api/internal/jwt"
	"github.com/ukane-philemon/labtracka-api/internal/otp"
	"github.com/ukane-philemon/labtracka-api/internal/request"
	"github.com/ukane-philemon/labtracka-api/internal/validator"
)

// handleSendOTP handles the "POST /otp" endpoint and sends an otp to the
// provided receiver.
func (s *Server) handleSendOTP(res http.ResponseWriter, req *http.Request) {
	var reqBody *sendOTPRequest
	err := request.DecodeJSONStrict(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	if validator.AnyValueEmpty(reqBody.Receiver, reqBody.DeviceID) {
		s.badRequest(res, req, "missing required field(s)")
		return
	}

	// TODO: Support sending OTP to phone number.
	if !validator.IsEmail(reqBody.Receiver) {
		s.badRequest(res, req, "a valid email is required to send OTP")
		return
	}

	if secs := s.optManager.SecsTillCanResendOTP(reqBody.DeviceID, reqBody.Receiver); secs > 0 {
		s.badRequest(res, req, fmt.Sprintf("you have previously requested for an OTP, please wait for %d seconds to request another one", secs))
		return
	}

	s.doInBackground("optManager.SendOTP", func() error {
		return s.optManager.SendOTP(reqBody.DeviceID, reqBody.Receiver, s.emailOtpSender)
	})

	s.sendSuccessResponse(res, req, "OTP sent successfully")
}

// handleOTPValidation handles the "POST /opt-validation" endpoint and checks
// that the provided otp is correct. A "validation_token" will be returned if
// successful validated.
func (s *Server) handleOTPValidation(res http.ResponseWriter, req *http.Request) {
	var reqBody *validateOTPRequest
	err := request.DecodeJSONStrict(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	if validator.AnyValueEmpty(reqBody.Receiver, reqBody.DeviceID, reqBody.OTP) {
		s.badRequest(res, req, "missing required field(s)")
		return
	}

	// TODO: Support sending OTP to phone number.
	if !validator.IsEmail(reqBody.Receiver) {
		s.badRequest(res, req, "a valid email is required to validate OTP")
		return
	}

	if len(reqBody.OTP) != otp.MaxLength {
		s.badRequest(res, req, "invalid OTP")
		return
	}

	validationToken, err := s.optManager.ValidateOTP(reqBody.DeviceID, reqBody.OTP, reqBody.Receiver)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("optManager.ValidateOTP error: %w", err))
		return
	}

	if validationToken == "" {
		s.badRequest(res, req, "invalid OTP")
		return
	}

	s.sendSuccessResponseWithData(res, req, map[string]string{
		"validation_token": validationToken,
		"message":          "OTP validated successfully",
	})
}

// handleLogin handles the "POST /login" endpoint and log a user with correct
// details into their account.
func (s *Server) handleLogin(res http.ResponseWriter, req *http.Request) {
	var reqBody *loginRequest
	err := request.DecodeJSONStrict(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	if validator.AnyValueEmpty(reqBody.DeviceID, reqBody.Email, reqBody.Password) {
		s.badRequest(res, req, "missing required field(s)")
		return
	}

	if !validator.IsEmail(reqBody.Email) || !validator.IsPasswordValid(reqBody.Password) {
		s.badRequest(res, req, "invalid email or password")
		return
	}

	var saveNewDeviceID bool
	if reqBody.EmailValidationToken != "" {
		if isValid := s.optManager.ValidateOTPValidationToken(reqBody.DeviceID, reqBody.EmailValidationToken, reqBody.Email); !isValid {
			s.badRequest(res, req, "invalid OTP")
			return
		}
		saveNewDeviceID = true
	}

	loginReq := &db.LoginRequest{
		Email:             reqBody.Email,
		Password:          reqBody.Password,
		ClientIP:          clientIP(req),
		DeviceID:          reqBody.DeviceID,
		SaveNewDeviceID:   saveNewDeviceID,
		NotificationToken: reqBody.NotificationToken,
	}

	adminInfo, err := s.db.LoginAdmin(loginReq)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, "incorrect email or password")
		} else if errors.Is(err, db.ErrorOTPRequired) {
			s.badRequest(res, req, "otp required")
		} else {
			s.serverError(res, req, fmt.Errorf("db.LoginAdmin: %w", err))
		}
		return
	}

	if saveNewDeviceID {
		s.optManager.DeleteOTPRecord(loginReq.ClientIP)
	}

	accessToken, err := s.jwtManager.GenerateJWtToken(reqBody.Email)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("jwtManager.GenerateJWtToken: %w", err))
		return
	}

	resp := &loginResponse{
		Admin: adminInfo,
		Auth: &authResponse{
			AccessToken:     accessToken,
			ExpiryInSeconds: uint64(jwt.JWTExpiry.Seconds()),
		},
	}

	if adminInfo.ServerAdmin || len(adminInfo.LabIDs) > 0 {
		resp.AdminStats, err = s.patientDB.AdminStats(adminInfo.LabIDs...)
		if err != nil {
			s.logger.Error("patientDB.AdminStats error: %v", err)
		}
	}

	if adminInfo.ServerAdmin || adminInfo.SuperAdmin {
		resp.AdminLabs, err = s.db.AdminLabs(adminInfo.Email)
		if err != nil {
			s.logger.Error("db.AdminLabs error: %v", err)
		}
	}

	s.sendSuccessResponseWithData(res, req, resp)
}

// handleResetPassword handles the "POST /reset-password" and resets the
// password of an existing admin.
func (s *Server) handleResetPassword(res http.ResponseWriter, req *http.Request) {
	var reqBody *resetPasswordRequest
	err := request.DecodeJSONStrict(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	if validator.AnyValueEmpty(reqBody.Email, reqBody.DeviceID, reqBody.NewPassword, reqBody.EmailValidationToken) {
		s.badRequest(res, req, "missing required field(s)")
		return
	}

	if !validator.IsEmail(reqBody.Email) {
		s.badRequest(res, req, "invalid email")
		return
	}

	if !validator.IsPasswordValid(reqBody.NewPassword) {
		s.badRequest(res, req, validator.PassWordErrorMsg)
		return
	}

	if isValid := s.optManager.ValidateOTPValidationToken(reqBody.DeviceID, reqBody.EmailValidationToken, reqBody.Email); !isValid {
		s.badRequest(res, req, "invalid OTP")
		return
	}

	err = s.db.ResetPassword(reqBody.Email, reqBody.NewPassword)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.ResetPassword error: %w", err))
		}
		return
	}

	s.optManager.DeleteOTPRecord(reqBody.DeviceID)

	s.sendSuccessResponse(res, req, "Password reset was successful, please proceed to login")
}

// handleRefreshAuthToken handles the "GET /refresh-auth-token" endpoint and
// returns a new access token for a logged in patient.
func (s *Server) handleRefreshAuthToken(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	accessToken, err := s.jwtManager.GenerateJWtToken(authID)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("jwtManager.GenerateJWtToken: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, &authResponse{
		AccessToken:     accessToken,
		ExpiryInSeconds: uint64(jwt.JWTExpiry.Seconds()),
	})
}

// handleChangePassword handles the "POST /change-password" endpoint and updates
// the password of an existing patient.
func (s *Server) handleChangePassword(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	var reqBody *changePasswordRequest
	err := request.DecodeJSONStrict(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, err.Error())
		return
	}

	if !validator.IsPasswordValid(reqBody.CurrentPassword) || !validator.IsPasswordValid(reqBody.NewPassword) {
		s.badRequest(res, req, validator.PassWordErrorMsg)
		return
	}

	err = s.db.ChangePassword(authID, reqBody.CurrentPassword, reqBody.NewPassword)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.ChangePassword error: %w", err))
		}
		return
	}

	s.sendSuccessResponse(res, req, "Password update was successful")
}
