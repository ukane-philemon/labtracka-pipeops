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
	// ResetPassword reset the password of an existing admin. Returns an
	// ErrorInvalidRequest if the email is not tied to an existing admin.
	ResetPassword(email, password string) error
	// ChangePassword updates the password for an existing admin. Returns an
	// ErrorInvalidRequest if email is not tied to an existing admin or current
	// password is incorrect.
	ChangePassword(email, currentPassword, newPassword string) error
	// Notifications returns all the notifications for patient sorted by unread
	// first.
	Notifications(email string) ([]*db.Notification, error)
	// MarkNotificationsAsRead marks the notifications with the provided noteIDs
	// as read.
	MarkNotificationsAsRead(email string, noteIDs ...string) error
	// Faqs returns information about frequently asked questions and help links.
	Faqs() (*db.Faqs, error)
	// UpdateFaqs updates the faqs in the database. This is a super admin only
	// feature.
	UpdateFaqs(faq *db.Faqs) (*db.Faqs, error)
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
