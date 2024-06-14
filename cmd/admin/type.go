package admin

import "time"

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
