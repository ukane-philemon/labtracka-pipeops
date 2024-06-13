package paystack

import (
	"encoding/json"
)

type Event struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// cardAuthorization represents Paystack's card authorization object. See
// https://paystack.com/docs/payments/recurring-charges#get-the-card-authorization
type CardAuthorization struct {
	AuthorizationCode string `json:"authorization_code,omitempty"`
	Bin               string `json:"bin,omitempty"`
	Last4             string `json:"last4,omitempty"`
	ExpMonth          string `json:"exp_month,omitempty"`
	ExpYear           string `json:"exp_year,omitempty"`
	CardType          string `json:"card_type,omitempty"`
	Bank              string `json:"bank,omitempty"`
	AccountName       string `json:"account_name,omitempty"`
	Reusable          bool   `json:"reusable,omitempty"`
	Signature         string `json:"signature,omitempty"`
}

// ChargeWebhookEvent is the event sent in response to bank card debit from
// Paystack.
type ChargeWebhookEvent struct {
	Status        string            `json:"status"`
	Reference     string            `json:"reference"`
	Amount        float64           `json:"amount"`
	Authorization CardAuthorization `json:"authorization"`
	PaidAt        string            `json:"paid_at"`
	Customer      struct {
		Email string `json:"email,omitempty"`
	} `json:"customer"`
	Metadata *map[string]interface{} `json:"metadata"`
}

// ChargeOptions are contains information to bind a card.
type ChargeOptions struct {
	Email        string
	AmountInKobo float64
	MetaData     map[string]interface{}
}

// Config is information required to send request to Paystack API.
type Config struct {
	BaseURL   string `long:"url" description:"Url for connecting with Paystack services"`
	AuthToken string `long:"authtoken" description:"AuthToken for authorizing Paystack API calls"`
}
