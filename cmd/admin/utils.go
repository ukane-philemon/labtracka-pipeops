package admin

import (
	"fmt"

	"github.com/ukane-philemon/labtracka-api/internal/otp"
)

func (s *Server) doInBackground(action string, fn func() error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := fn()
		if err != nil {
			s.logger.Error("%s: %v", action, err)
		}
	}()
}

// emailOtpSender sends an otp to the email entity provided.
func (s *Server) emailOtpSender(entity string, otpInfo *otp.TimedValue) error {
	data := map[string]string{
		"OTP":    otpInfo.Value,
		"Expiry": fmt.Sprintf("%d minutes", int(otpInfo.Expiry.Minutes())),
	}

	err := s.mailer.Send(entity, data, "otp.tmpl")
	if err != nil {
		return fmt.Errorf("mailer.Send error: %w", err)
	}

	return nil
}
