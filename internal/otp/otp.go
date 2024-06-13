package otp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"
)

const (
	// OTPExpiry is the maximum duration during which an OTP or
	// otpValidationToken is valid. Any OTP or otpValidationToken is considered
	// expired 10 minutes after it was generated.
	OTPExpiry = time.Minute * 10

	// TimeBetweenOTPResends is the minimum duration that must pass before a
	// previously-sent OTP can be resent.
	TimeBetweenOTPResends = 1 * time.Minute

	MaxLength = 4
)

type OTPValidationInfo struct {
	ValidateEntity  string
	OTP             TimedValue
	ValidationToken TimedValue
}

type TimedValue struct {
	Value  string
	Time   time.Time
	Expiry time.Duration
}

// IsExpired checks if the Time of TimedValue is expired.
func (t *TimedValue) IsExpired() bool {
	return time.Since(t.Time) > t.Expiry
}

type Manager struct {
	devMode          bool
	mtx              sync.RWMutex
	otpValidationMap map[string]*OTPValidationInfo
}

func NewManager(devMode bool) *Manager {
	return &Manager{
		devMode:          devMode,
		otpValidationMap: make(map[string]*OTPValidationInfo),
	}
}

// SendOTP generates an OTP for the validateEntity provided and sends it using
// the otpSender provided which could be an email OTP sender or a phone number
// OTP sender. The validateEntity is the data to be validated and may be an
// email address or phone number.
func (om *Manager) SendOTP(deviceID, validateEntity string, otpSender func(entity string, otpInfo *TimedValue) error) error {
	otp, err := RandomOTP()
	if err != nil {
		return fmt.Errorf("RandomOTP error: %w", err)
	}

	if om.devMode {
		otp = "123456"
	}

	otpInfo := &TimedValue{Value: otp, Time: time.Now(), Expiry: OTPExpiry}

	if err := otpSender(validateEntity, otpInfo); err != nil {
		return err
	}

	om.RecordOTPForDevice(deviceID, &OTPValidationInfo{
		ValidateEntity: validateEntity,
		OTP:            TimedValue{Value: otp, Time: time.Now(), Expiry: OTPExpiry},
	})
	return nil
}

// RecordOTPForDevice saves an OTP record for the provided entity against the
// provided deviceID. Overrides any previously recorded OTP for the deviceID.
func (om *Manager) RecordOTPForDevice(deviceID string, info *OTPValidationInfo) {
	om.mtx.Lock()
	om.otpValidationMap[deviceID] = info
	om.mtx.Unlock()
}

// SecsTillCanResendOTP checks if the deviceID provided has an un-validated otp
// for the specified entity and returns the minimum number of seconds that must
// pass before the caller may send another OTP request for this same entity. If
// 0 seconds is returned, the caller may send another OTP request right away.
// NOTE: This doesn't care whether the caller has another OTP info tied to the
// same deviceID but for a different entity, or if the caller has an OTP for
// this entity which has been validated with a validation token. If such OTP
// info exists, it may be overriden. This is intended to discourage callers from
// performing multiple OTP validations from the same device at the same time.
func (om *Manager) SecsTillCanResendOTP(deviceID, entity string) int64 {
	om.mtx.Lock()
	defer om.mtx.Unlock()
	auth, found := om.otpValidationMap[deviceID]
	if !found || auth.ValidateEntity != entity || auth.ValidationToken.Value != "" ||
		time.Since(auth.OTP.Time) > TimeBetweenOTPResends {
		return 0
	}
	return int64(math.Ceil((TimeBetweenOTPResends - time.Since(auth.OTP.Time)).Seconds()))
}

// IsValidOTP validates the otp against the validateEntity using the deviceID.
// The validateEntity is the data to be validated and may be an email address or
// phone number.
// NOTE: This will continue to return true for a valid OTP until it expires or a
// validationToken is saved for the deviceID. Callers should delete as
// appropriate via om.DeleteOTPRecord.
func (om *Manager) IsValidOTP(deviceID, otp, validateEntity string) bool {
	om.mtx.Lock()
	defer om.mtx.Unlock()
	return om.isValidOTP(deviceID, otp, validateEntity)
}

// isValidOTP validates an otp info without locking the otpMtx. Mutex must be
// locked before this method is called.
func (om *Manager) isValidOTP(deviceID, otp, validateEntity string) bool {
	auth, found := om.otpValidationMap[deviceID]
	if !found || auth.OTP.IsExpired() { // return early if expired, no need to check anything.
		delete(om.otpValidationMap, deviceID)
		return false
	}

	// Must not have been previously validated (auth.validationToken.value must be "")
	return auth.ValidationToken.Value == "" && auth.OTP.Value == otp && auth.ValidateEntity == validateEntity
}

// ValidateOTP is similar to IsValidOTP but also generates and returns an OTP
// validation token which can be used to verify that the caller had previously
// validated the OTP if the OTP is valid.
func (om *Manager) ValidateOTP(deviceID, otp, validateEntity string) (string, error) {
	om.mtx.Lock()
	defer om.mtx.Unlock()

	if !om.isValidOTP(deviceID, otp, validateEntity) {
		return "", nil
	}

	token, err := RandomToken(32)
	if err != nil {
		return "", fmt.Errorf("RandomToken error: %v", err)
	}

	om.otpValidationMap[deviceID].ValidationToken = TimedValue{
		Value:  token,
		Time:   time.Now(),
		Expiry: OTPExpiry,
	}
	return token, nil
}

// ValidateOTPValidationToken validates the provided otpValidationToken using
// the given deviceID. NOTE: This will continue to return true for a valid token
// until it expires. Callers should delete as appropriate via om.DeleteOTPRecord.
func (om *Manager) ValidateOTPValidationToken(deviceID, otpValidationToken, entity string) bool {
	om.mtx.Lock()
	defer om.mtx.Unlock()
	auth, found := om.otpValidationMap[deviceID]
	if !found || auth.ValidationToken.Value != otpValidationToken || auth.ValidateEntity != entity {
		return false
	}
	if auth.ValidationToken.IsExpired() {
		delete(om.otpValidationMap, deviceID)
		return false
	}
	return true
}

func (om *Manager) DeleteOTPRecord(deviceID string) {
	om.mtx.Lock()
	defer om.mtx.Unlock()
	delete(om.otpValidationMap, deviceID)
}

// RandomOTP generates and returns a random six-digit OTP.
func RandomOTP() (string, error) {
	b := make([]byte, MaxLength) // each value will be a number between 0 and 255, if single digits are returned, we'll get 6 digits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d%d%d%d", b[0], b[1], b[2], b[3])[:MaxLength], nil
}

// RandomToken generates a hex-encoded random token with length*2 characters.
// Authentication tokens should pass a length >= 16 to get >=32 char strings.
func RandomToken(length int) (string, error) {
	b, err := RandomBytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RandomBytes generates and returns a random byte slice of the specified
// length.
func RandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
