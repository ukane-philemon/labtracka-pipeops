package admin

import (
	"fmt"
	"net/http"

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
