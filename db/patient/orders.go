package patient

import "github.com/ukane-philemon/labtracka-api/db"

// AddSubAccount adds a new sub account to a patient's profile.
func (m *MongoDB) AddSubAccount(email string, account *db.SubAccount) ([]*db.SubAccountInfo, error) {
	return nil, nil
}

// SubAccounts returns the sub account for the customer with the provided
// email address.
func (m *MongoDB) SubAccounts(email string) ([]*db.SubAccountInfo, error) {
	return nil, nil
}

// RemoveSubAccount removes a sub account from a patient's record and
// returns the remaining sub accounts. Return db.ErrorInvalidRequest if
// subAccountID does not exist.
func (m *MongoDB) RemoveSubAccount(email, subAccountID string) ([]*db.SubAccountInfo, error) {
	return nil, nil
}

// AddNewAddress adds a new address to a patient's profile.
func (m *MongoDB) AddNewAddress(email string, address *db.CustomerAddress) ([]*db.CustomerAddress, error) {
	return nil, nil
}

// PatientOrders returns a list of orders made by the patient with the
// provided email.
func (m *MongoDB) PatientOrders(email string) ([]*db.Order, error) {
	return nil, nil
}

// CreatePatientOrder creates a new order for the patient and returns the
// orderID and amount after validating the order.
func (m *MongoDB) CreatePatientOrder(email string, orderReq *db.CreateOrderRequest) (string, float64, error) {
	return "", 0, nil
}

// UpdatePatientOrder updates the status for a patient order.
func (m *MongoDB) UpdatePatientOrder(email string, orderID, status string) error {
	return nil
}
