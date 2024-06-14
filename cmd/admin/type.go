package admin

import (
	"time"

	"github.com/ukane-philemon/labtracka-api/db"
)

type sendOTPRequest struct {
	DeviceID string `json:"device_id"`
	Receiver string `json:"receiver"`
}

type validateOTPRequest struct {
	DeviceID string `json:"device_id"`
	Receiver string `json:"receiver"`
	OTP      string `json:"otp"`
}

type timedValue struct {
	value  string
	expiry time.Time
}

// loginRequest is information required for login.
type loginRequest struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	DeviceID             string `json:"device_id"`
	NotificationToken    string `json:"notification_token"`     // TODO: Validate
	EmailValidationToken string `json:"email_validation_token"` // optional
}

type loginResponse struct {
	*db.Admin
	db.AdminStats
	AdminLabs []*db.AdminLabInfo `json:"admin_labs"` // only for super amin
	Auth      *authResponse      `json:"auth"`
}

type authResponse struct {
	AccessToken     string `json:"access_token"`
	ExpiryInSeconds uint64 `json:"expiry_in_seconds"`
}