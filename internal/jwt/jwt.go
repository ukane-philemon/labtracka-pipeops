package jwt

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cristalhq/jwt"
)

const (
	jwtIssuer = "LabTracka Ltd"

	JWTExpiry           = 15 * time.Minute
	jwtAudienceAdmin    = "Admin"
	jwtAudienceCustomer = "Customer"
)

type Manager struct {
	aud     string
	signer  jwt.Signer
	pubKey  ed25519.PublicKey
	privKey ed25519.PrivateKey

	builder   *jwt.TokenBuilder
	validator *jwt.Validator
}

// NewJWTManager returns a new manager for jwt tokens.
func NewJWTManager(forAdmin bool) (*Manager, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("ed25519.GenerateKey error: %w", err)
	}

	signer, err := jwt.NewEdDSA(publicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("jwt.NewEdDSA error: %w", err)
	}

	m := &Manager{
		aud:     jwtAudienceCustomer,
		signer:  signer,
		pubKey:  publicKey,
		privKey: privateKey,
		builder: jwt.NewTokenBuilder(signer),
	}

	if forAdmin {
		m.aud = jwtAudienceAdmin
	}
	m.validator = jwt.NewValidator(jwt.AudienceChecker(jwt.Audience{m.aud}), jwt.ValidAtNowChecker(), jwt.IssuerChecker(jwtIssuer), jwt.ValidAtNowChecker())

	return m, nil
}

// GenerateJWtToken generates a new jwt token for the specified id.
func (m *Manager) GenerateJWtToken(id string) (string, error) {
	claims := jwt.StandardClaims{
		Audience:  []string{m.aud},
		ExpiresAt: jwt.Timestamp(time.Now().Add(JWTExpiry).Unix()),
		ID:        id,
		IssuedAt:  jwt.Timestamp(time.Now().Unix()),
		Issuer:    jwtIssuer,
	}

	token, err := m.builder.Build(claims)
	if err != nil {
		return "", fmt.Errorf("m.builder.Build error: %w", err)
	}

	return token.String(), nil
}

// IsValidToken checks that the provided token is valid and returns the unique
// id added to the auth token.
func (m *Manager) IsValidToken(jwtToken string) (string, bool) {
	token, err := jwt.ParseAndVerifyString(jwtToken, m.signer)
	if err != nil {
		return "", false
	}

	var claims *jwt.StandardClaims
	err = json.Unmarshal(token.RawClaims(), &claims)
	if err != nil {
		return "", false
	}

	err = m.validator.Validate(claims)
	if err != nil {
		return "", false
	}

	return claims.ID, true
}
