package patient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ukane-philemon/labtracka-api/internal/funcs"
	"github.com/ukane-philemon/labtracka-api/internal/paystack"
)

// handlePaystackWebhookEvent handles the webhook events that are sent from
// Paystack for transaction updates.
func (r *Server) handlePaystackWebhookEvent(res http.ResponseWriter, req *http.Request) {
	funcName := "handlePaystackWebhookEvent"
	paystackSig := req.Header.Get("x-paystack-signature")
	if paystackSig == "" {
		r.badRequest(res, req, "received webhook notification without Paystack signature")
		return
	}

	body, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		r.badRequest(res, req, fmt.Sprintf("%s: invalid request body error: %v", funcName, err))
		return
	}

	// Verify paystack signature.
	if !r.paystack.VerifyRequest(paystackSig, body) {
		r.badRequest(res, req, "verifiable paystack signature is required to access this webhook endpoint")
		return
	}

	e := new(paystack.Event)
	err = json.Unmarshal(body, e)
	if err != nil {
		r.serverError(res, req, fmt.Errorf("%s: body: %s: json.Unmarshal error: %v",
			funcName, string(body), err))
		return
	}

	// Check event type and handle accordingly.
	if strings.Contains(paystack.SuccessChargeEvents, e.Event) {
		event := new(paystack.ChargeWebhookEvent)
		err := json.Unmarshal(e.Data, event)
		if err != nil {
			r.serverError(res, req, fmt.Errorf("%s: body: %s: json.Unmarshal error: %v",
				funcName, string(e.Data), err))
			return
		}

		var metaData map[string]interface{}
		if event.Metadata != nil {
			metaData = *event.Metadata
		}

		// Retrieve metadata.
		email := metaData[emailIDKeyForTx].(string)
		orderID := metaData[orderIDKeyForTx].(string)

		title := "Sorry, your order failed"
		if strings.EqualFold(event.Status, "success") {
			title = "Your order was successful!"
			err = r.db.UpdatePatientOrder(email, orderID, "Paid")
			if err != nil {
				r.serverError(res, req, fmt.Errorf("db.UpdatePatientOrder error: %w", err))
				return
			}
		}

		// TODO: Send in app and email notification for the order.
		r.mailer.Send(email, map[string]any{
			"Title":  title,
			"Amount": funcs.KoboToNaira(event.Amount),
		}, "payment.tmpl")

		fmt.Printf("\nFailed payment tx: %+v", event) // TODO: Remove
	}

	// Respond to the webhook event first after successfully reading the request
	// body.
	r.sendSuccessResponse(res, req, "")
}
