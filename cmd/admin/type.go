package admin

type sendOTPRequest struct {
	DeviceID string `json:"device_id"`
	Receiver string `json:"receiver"`
}
