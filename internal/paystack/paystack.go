package paystack

import (
	"context"
	"crypto"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/ukane-philemon/labtracka-api/internal/funcs"
	"github.com/ukane-philemon/labtracka-api/internal/request"
	"github.com/ukane-philemon/labtracka-api/internal/validator"
)

const (
	defaultPaystackURL = "https://api.paystack.co"

	// SuccessChargeEvents is the webhook event that is sent when a card has been
	// successfully debited.
	SuccessChargeEvents = "charge.success transfer.success"

	// labTrackaTx is the key used to identify LabTracka transactions.
	labTrackaTx = "labtracka_transaction"
	appName     = "LabTracka"
)

var (
	// This is a special Paystack error used to identify when a bank card
	// requires 2FA.
	ErrorCardChallenged = errors.New("Card requires authorization")
	ErrorInvalidRequest = errors.New("Invalid Request")
)

// DefaultConfig returns an instance of the Paystack Config with default values
// set.
func DefaultConfig() Config {
	return Config{
		BaseURL: defaultPaystackURL,
	}
}

// Client communicates with the Client API endpoints.
type Client struct {
	ctx context.Context
	cfg Config
}

// New returns a new instance of *Client.
func New(ctx context.Context, cfg Config, devMode bool) (*Client, error) {
	if cfg.AuthToken == "" {
		return nil, errors.New("no auth token provided")
	}

	_, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("error validating BaseURL: %w", err)
	}

	c := &Client{
		ctx: ctx,
		cfg: cfg,
	}

	if err := c.validateOperationMode(devMode); err != nil {
		return nil, fmt.Errorf("error validating operation mode: %w", err)
	}

	return c, nil
}

// validateOperationMode checks operation mode with a heuristic to tell if we
// are using the correct auth token for the current operation mode and the auth
// token is valid. This is not a foolproof method, but it's better than nothing.
func (c *Client) validateOperationMode(devMode bool) error {
	// Create a new customer to test the auth token and check the response for
	// the operation mode.
	testEmail := "labtracka@test.com"
	reqBody := map[string]interface{}{
		"email":      testEmail,
		"first_name": appName,
		"last_name":  "Server",
	}

	resp := &struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Domain string `json:"domain"`
		} `json:"data"`
	}{}
	err := c.callPaystackAPI("POST", "customer", reqBody, resp)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "exist") {
		// Test customer already exists, so fetch the customer info. Test mode
		// never error if customer already exists but this is a fallback.
		err = c.callPaystackAPI("GET", fmt.Sprintf("customer/%s", testEmail), nil, resp)
	}
	if err != nil {
		return fmt.Errorf("error validating Paystack auth token: %w", err)
	}

	if strings.Contains(strings.ToLower(resp.Data.Domain), "test") != devMode {
		return fmt.Errorf("invalid auth token for operation mode (DevMode == %v)", devMode)
	}

	return nil
}

// Ready checks that a Paystack auth token is set.
func (c *Client) Ready() bool {
	return c.cfg.AuthToken != ""
}

// VerifyRequest verifies that the provided requestBody was signed using
// Paystack's auth token.
func (c *Client) VerifyRequest(signature string, requestBody []byte) bool {
	pHash := hmac.New(crypto.SHA512.New, []byte(c.cfg.AuthToken))
	pHash.Write(requestBody)
	hash := pHash.Sum(nil)
	b := make([]byte, hex.EncodedLen(len(hash)))
	hex.Encode(b, hash)
	return hmac.Equal(b, []byte(signature))
}

// Charge starts the process to charge an account and returns an Authorization
// URL for callers to finish the process. A default narration is used for the
// transaction.
func (c *Client) Charge(opts *ChargeOptions) (string, error) {
	if !validator.IsEmail(opts.Email) {
		return "", fmt.Errorf("invalid email: %s", opts.Email)
	}

	reqBody := map[string]interface{}{
		"email":    opts.Email,
		"amount":   fmt.Sprintf("%f", opts.AmountInKobo),
		"currency": "NGN",                                     // TODO: support other currencies
		"channels": []string{"card", "bank", "bank_transfer"}, // TODO: Bind card
	}

	if opts.MetaData == nil {
		opts.MetaData = make(map[string]interface{}, 0)
	}

	opts.MetaData[labTrackaTx] = true
	reqBody["metadata"] = opts.MetaData

	resp := new(struct {
		Ok      bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
		} `json:"data"`
	})

	err := c.callPaystackAPI("POST", "transaction/initialize", reqBody, resp)
	if err != nil {
		return "", fmt.Errorf("paystack api error: %w", err)
	}

	if !resp.Ok {
		return "", errors.New(resp.Message)
	}

	return resp.Data.AuthorizationURL, nil
}

// DeactivateCardAuthorization deactivates a bank cards's authorization.
func (c *Client) DeactivateCardAuthorization(authCode string) error {
	if authCode == "" {
		return errors.New("deleted bank card authorization code is empty")
	}

	reqBody := map[string]string{
		"authorization_code": authCode,
	}

	resp := new(struct {
		Ok bool `json:"status"`
	})

	err := c.callPaystackAPI("POST", "customer/deactivate_authorization", reqBody, resp)
	if err != nil {
		return fmt.Errorf("callPaystackAPI error: %w", err)
	}

	if !resp.Ok {
		return errors.New("invalid bank card authorization")
	}

	return nil
}

// RefundTransaction refunds the full amount for the transaction identified with
// the txReference provided.
func (c *Client) RefundTransaction(txReference, message string) error {
	reqBody := map[string]string{
		"transaction":   txReference,
		"merchant_note": message,
	}

	resp := new(struct {
		Ok      bool   `json:"status"`
		Message string `json:"message"`
	})

	err := c.callPaystackAPI("POST", "refund", reqBody, resp)
	if err != nil {
		return fmt.Errorf("callPaystackAPI error: %w", err)
	}

	if !resp.Ok {
		return fmt.Errorf("refund error: %v", resp.Message)
	}

	return nil
}

// ChargeBankCard debits a bank card using the authorization code provided.
// Returns reference, amount debited, fee and error.
func (c *Client) ChargeBankCard(partial bool, amount float64, email, authorizationCode, narration string) (string, float64, float64, error) {
	reqBody := map[string]interface{}{
		"email":              email,
		"amount":             funcs.NairaToKobo(amount),
		"authorization_code": authorizationCode,
		"currency":           "NGN",
		"metadata": map[string]interface{}{
			"narration": narration,
			labTrackaTx: true,
		},
	}

	endpoint := "transaction/charge_authorization"

	// Check if partial debit is enabled if partial is specified unless default
	// to normal debit.
	if partial {
		// TODO: we could specify an optional at_least amount in the request
		// body and also handle scenarios where we received an error that
		// partial debit is not activated but server says otherwise but let's
		// leave that for now.
		endpoint = "transaction/partial_debit"
		reqBody["at_least"] = 100 // naira
	}

	resp := new(struct {
		Ok      bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Paused    bool    `json:"paused"`
			Status    string  `json:"status"`
			Reference string  `json:"reference"`
			Message   string  `json:"message"`
			Amount    float64 `json:"amount"` // actual amount debited.
			Fees      float64 `json:"fees"`
		} `json:"data"`
	})

	err := c.callPaystackAPI("POST", endpoint, reqBody, resp)
	if err != nil {
		return "", 0, 0, fmt.Errorf("callPaystackAPI error: %w", err)
	}

	if resp.Data.Paused {
		// Card has been challenged by customer's bank.
		return "", 0, 0, ErrorCardChallenged
	}

	if !resp.Ok {
		return "", 0, 0, fmt.Errorf("error debiting bank card: %s", resp.Message)
	}

	if !strings.EqualFold(resp.Data.Status, "success") {
		return "", 0, 0, fmt.Errorf("%w: %s", ErrorInvalidRequest, resp.Data.Message)
	}

	return resp.Data.Reference, funcs.KoboToNaira(resp.Data.Amount), funcs.KoboToNaira(resp.Data.Fees), nil
}

// callPaystackAPI makes an API request to the provided paystack endpoint and
// returns the response code and bytes.
func (c *Client) callPaystackAPI(method string, endpoint string, req interface{}, response interface{}) error {
	_, respBody, err := request.CallAPI(c.ctx, method, fmt.Sprintf("%s/%s", c.cfg.BaseURL, endpoint), c.cfg.AuthToken, req)
	if err != nil {
		return fmt.Errorf("commons.CallAPI error: %w", err)
	}

	err = json.Unmarshal(respBody, response)
	if err != nil {
		return fmt.Errorf("json.Unmarshal error: %v", err)
	}

	return nil
}
