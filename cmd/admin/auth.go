package admin

import (
	"fmt"
	"net/http"

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
