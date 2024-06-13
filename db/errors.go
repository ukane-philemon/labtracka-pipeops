package db

import "errors"

var (
	ErrorInvalidRequest = errors.New("invalid request")
	ErrorOTPRequired    = errors.New("otp required")
)
