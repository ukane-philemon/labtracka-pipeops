package admin

import "github.com/ukane-philemon/labtracka-api/db"

// LoginAdmin logs an admin into their account. Returns an ErrorInvalidRequest
// is user email or password is invalid/not correct or does not exist or an
// ErrorOTPRequired if otp validation is required for this account.
func (m *MongoDB) LoginAdmin(req *db.LoginRequest) (*db.Admin, error) {
	return nil, nil
}

// ResetPassword reset the password of an existing admin. Returns an
// ErrorInvalidRequest if the email is not tied to an existing admin.
func (m *MongoDB) ResetPassword(email, password string) error {
	return nil
}
