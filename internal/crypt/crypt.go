package crypt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// pubKeyValidationExpiry is the time it takes before a public key validation
// becomes invalid.
const pubKeyValidationExpiry = 2 * time.Hour

type pubKeyValidation struct {
	key    *rsa.PublicKey
	expiry time.Time
}

type PublicKeyValidator struct {
	mtx      sync.RWMutex
	messages map[string]*pubKeyValidation
}

func NewPublicKeyValidator() *PublicKeyValidator {
	return &PublicKeyValidator{
		messages: make(map[string]*pubKeyValidation),
	}
}

// EncryptWithPublicKey generates a random token and encrypts it with the
// provided pubKey. Returns the resulting hex encoded cipher text.
func (pv *PublicKeyValidator) EncryptWithPublicKey(token string, pubKey *rsa.PublicKey) (string, int64, error) {
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, []byte(token), nil)
	if err != nil {
		return "", 0, err
	}

	expiry := time.Now().Add(pubKeyValidationExpiry)
	pv.mtx.Lock()
	pv.messages[token] = &pubKeyValidation{
		key:    pubKey,
		expiry: expiry,
	}
	pv.mtx.Unlock()

	return hex.EncodeToString(encrypted), expiry.Unix(), nil
}

// IsValidPublicKeyToken checks that the decrypted cipher text returned by
// me.EncryptWithPublicKey is correct and valid. The provided public key must
// match the public key tied to the provided token.
func (pv *PublicKeyValidator) IsValidPublicKeyToken(token string, pubKey *rsa.PublicKey) bool {
	pv.mtx.RLock()
	defer pv.mtx.RUnlock()
	pubKeyInfo, found := pv.messages[token]
	sameKey := pubKey.Equal(pubKeyInfo.key)
	notExpired := time.Now().Before(pubKeyInfo.expiry)
	return found && sameKey && notExpired
}

// DeleteToken removes an already validated token. Call this method after
// successfully validating with IsValidPublicKeyToken.
func (pv *PublicKeyValidator) DeleteToken(token string) {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()
	delete(pv.messages, token)
}

// Clean removes expired pub key validation messages and should be run a a
// goroutine.
func (pv *PublicKeyValidator) Clean(ctx context.Context) {
	for {
		ticker := time.NewTicker(pubKeyValidationExpiry)
		defer ticker.Stop()

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			pv.mtx.Lock()
			for token, v := range pv.messages {
				if now.After(v.expiry) {
					delete(pv.messages, token)
				}
			}
			pv.mtx.Unlock()
		}
	}
}
