package admin

import "github.com/ukane-philemon/labtracka-api/db"

type Database interface {
	// LoginAdmin logs an admin into their account. Returns an
	// ErrorInvalidRequest is user email or password is invalid/not correct or
	// does not exist or an ErrorOTPRequired if otp validation is required for
	// this account.
	LoginAdmin(req *db.LoginRequest) (*db.Admin, error)
	// AdminLabs returns a list of labs added to the db for only super admin.
	// The provided email must match a super admin.
	AdminLabs(email string) ([]*db.AdminLabInfo, error)
	Shutdown()
}

type PatientDatabase interface {
	// Orders returns all the orders pending orders in the patient db for the
	// provided labIDs.
	Orders(labIDs ...string) (map[string]map[string][]*db.Order, error)
	// AdminStats returns some admin stats for display. If no lab id is
	// returned, all current stats will be returned.
	AdminStats(labIDs ...string) (db.AdminStats, error)
}
