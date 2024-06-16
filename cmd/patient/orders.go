package patient

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ukane-philemon/labtracka-api/db"
	"github.com/ukane-philemon/labtracka-api/internal/funcs"
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

	var reqBody *createOrderRequest
	err := request.DecodeJSON(res, req, &reqBody)
	if err != nil {
		s.badRequest(res, req, "invalid request body")
		return
	}

	if err := reqBody.PatientAddress.Validate(); err != nil {
		s.badRequest(res, req, err.Error())
		return
	}

	patient, err := s.db.PatientInfoWithID(reqBody.PatientID)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.PatientInfoWithID error: %w", err))
		return
	}

	if reqBody.SubAccountID != "" && !patient.HasSubAccount(reqBody.SubAccountID) {
		s.badRequest(res, req, "you cannot create an order for sub account")
		return
	}

	order := &db.OrderInfo{
		PatientID:    reqBody.PatientID,
		SubAccountID: reqBody.SubAccountID,
		Fee:          feesInNGN,
		Status:       db.OrderStatusPendingPayment,
		Address:      reqBody.PatientAddress,
	}

	for _, testID := range reqBody.Tests {
		testInfo, err := s.adminDB.LabTest(testID)
		if err != nil {
			if errors.Is(err, db.ErrorInvalidRequest) {

			} else {
				s.serverError(res, req, fmt.Errorf("adminDB.LabTest error: %w", err))
			}
			return
		}

		if !testInfo.IsActive {
			s.badRequest(res, req, fmt.Sprintf("%s has been disabled", testInfo.Name))
			return
		}

		order.Tests = append(order.Tests, &db.OrderTest{
			LabID:    testInfo.LabID,
			LabName:  testInfo.LabName,
			TestID:   testID,
			TestName: testInfo.Name,
			Amount:   testInfo.Price,
		})

		order.Description += ", " + testInfo.Name
		order.TotalAmount += testInfo.Price
	}

	order.Description = strings.Trim(order.Description, ",")

	// Create the order for the patient and generate a payment url for the
	// order. The order would be confirmed when we receive a payment webhook
	// from our payment provider.
	orderID, err := s.db.CreatePatientOrder(authID, order)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.CreatePatientOrder error: %w", err))
		}
		return
	}

	// Create a payment URL for order amount.
	paymentURL, err := s.paystack.Charge(&paystack.ChargeOptions{
		Email:        authID,
		AmountInKobo: funcs.NairaToKoboAmt(order.TotalAmount + order.Fee),
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
