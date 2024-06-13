package patient

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
	*db.Customer
	db.PatientStats
	AvailableLabs []*db.BasicLabInfo `json:"available_labs"`
	Auth          *authResponse      `json:"auth"`
}

// resetPasswordRequest is information required to reset customer password.
type resetPasswordRequest struct {
	Email                string `json:"email"`
	NewPassword          string `json:"new_password"`
	DeviceID             string `json:"device_id"`
	EmailValidationToken string `json:"email_validation_token"`
}

// changePasswordRequest is information required to change password for a logged
// in customer.
type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type createAccountRequest struct {
	db.CustomerInfo
	DeviceID             string `json:"device_id"`
	Password             string `json:"password"`
	EmailValidationToken string `json:"email_validation_code"`
	DecryptedToken       string `json:"decrypted_token"`
}

type authResponse struct {
	AccessToken     string `json:"access_token"`
	ExpiryInSeconds uint64 `json:"expiry_in_seconds"`
}

type patientProfile struct {
	db.CustomerInfo
}
