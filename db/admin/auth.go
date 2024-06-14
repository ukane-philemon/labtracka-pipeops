package admin

import "github.com/ukane-philemon/labtracka-api/db"

// LoginAdmin logs an admin into their account. Returns an ErrorInvalidRequest
// is user email or password is invalid/not correct or does not exist or an
// ErrorOTPRequired if otp validation is required for this account.
func (m *MongoDB) LoginAdmin(req *db.LoginRequest) (*db.Admin, error) {
	return nil, nil
}
