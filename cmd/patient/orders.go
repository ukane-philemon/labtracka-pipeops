package patient

import (
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/ukane-philemon/labtracka-api/db"
	"github.com/ukane-philemon/labtracka-api/internal/paystack"
	"github.com/ukane-philemon/labtracka-api/internal/request"
)

const (
	feesInNGN       float64 = 5000
	orderIDKeyForTx         = "orderID"
	emailIDKeyForTx         = "authID"
)

// handleCreateOrder handles the "POST /order" endpoint and creates a new order
// for the patient. Returns details to complete payment.
func (s *Server) handleCreateOrder(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	var reqBody *db.CreateOrderRequest
	err := request.DecodeJSON(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	// Create the order for the patient and generate a payment url for the
	// order. The order would be confirmed when we receive a payment webhook
	// from provider.
	orderID, orderAmount, err := s.db.CreatePatientOrder(authID, reqBody)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.CreatePatientOrder error: %w", err))
		}
		return
	}

	// Create a payment URL for order amount.
	totalOrderAmount := orderAmount + feesInNGN
	paymentURL, err := s.paystack.Charge(&paystack.ChargeOptions{
		Email:        authID,
		AmountInKobo: math.Ceil(totalOrderAmount * 100),
		MetaData: map[string]interface{}{
			orderIDKeyForTx: orderID,
			emailIDKeyForTx: authID,
		},
	})
	if err != nil {
		s.serverError(res, req, fmt.Errorf("paystack.Charge error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, map[string]any{
		"payment_url": paymentURL,
	})
}

// handleGetOrders handles the "GET /orders" endpoint and returns all current
// patient orders.
func (s *Server) handleGetOrders(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.authenticationRequired(res, req)
		return
	}

	orders, err := s.db.PatientOrders(authID)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.PatientOrders error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, orders)
}
