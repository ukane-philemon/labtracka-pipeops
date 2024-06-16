package patient

import "github.com/ukane-philemon/labtracka-api/db"

// PatientOrders returns a list of orders made by the patient with the
// provided email.
func (m *MongoDB) PatientOrders(email string) ([]*db.Order, error) {
	return nil, nil
}

// CreatePatientOrder creates a new order for the patient and returns the
// orderID and amount after validating the order.
func (m *MongoDB) CreatePatientOrder(email string, orderReq *db.OrderInfo) (string, error) {
	return "", nil
}

// UpdatePatientOrder updates the status for a patient order.
func (m *MongoDB) UpdatePatientOrder(email string, orderID, status string) error {
	return nil
}
