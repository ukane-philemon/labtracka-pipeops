package patient

import "github.com/ukane-philemon/labtracka-api/db"

type Database interface {
	Shutdown()

	/**** Patient ****/

	// CreateAccount creates a new customer and their information are saved to
	// the database. Returns an ErrorInvalidRequest if user email is already
	// tied to another customer.
	CreateAccount(req *db.CreateAccountRequest) error
	// LoginCustomer logs a customer into their account. Returns an
	// ErrorInvalidRequest is user email or password is invalid/not correct or
	// does not exist or an ErrorOTPRequired if otp validation is required for
	// this account.
	LoginCustomer(loginReq *db.LoginRequest) (*db.Customer, error)
	// ResetPassword reset the password of an existing customer. Returns an
	// ErrorInvalidRequest if the email is not tied to an existing customer.
	ResetPassword(email, password string) error
	// ChangePassword updates the password for an existing customer. Returns an
	// ErrorInvalidRequest if email is not tied to an existing customer or
	// current password is incorrect.
	ChangePassword(email, currentPassword, newPassword string) error
	// AddSubAccount adds a new sub account to a patient's profile.
	AddSubAccount(email string, account *db.SubAccount) ([]*db.SubAccountInfo, error)
	// SubAccounts returns the sub account for the customer with the provided
	// email address.
	SubAccounts(email string) ([]*db.SubAccountInfo, error)
	// RemoveSubAccount removes a sub account from a patient's record and
	// returns the remaining sub accounts. Return db.ErrorInvalidRequest if
	// subAccountID does not exist.
	RemoveSubAccount(email, subAccountID string) ([]*db.SubAccountInfo, error)
	// AddNewAddress adds a new address to a patient's profile.
	AddNewAddress(email string, address *db.CustomerAddress) ([]*db.CustomerAddress, error)
	// PatientOrders returns a list of orders made by the patient with the
	// provided email.
	PatientOrders(email string) ([]*db.Order, error)
	// CreatePatientOrder creates a new order for the patient and returns the
	// orderID and amount after validating the order.
	CreatePatientOrder(email string, orderReq *db.CreateOrderRequest) (string, float64, error)
	// UpdatePatientOrder updates the status for a patient order.
	UpdatePatientOrder(email string, orderID, status string) error
	// Notifications returns all the notifications for customer sorted by unread
	// first.
	Notifications(email string) ([]*db.Notification, error)
	// MarkNotificationsAsRead marks the notifications with the provided noteIDs
	// as read.
	MarkNotificationsAsRead(email string, noteIDs ...string) error
}

type AdminDatabase interface {
	/**** Patient ****/

	PatientLabStats(patientID string) (db.PatientStats, error)

	// Results returns all results for the customer with the specified email
	// address.
	Results(patientID string) ([]*db.LabResult, error)

	/**** Labs ****/

	// Labs returns a list of available labs.
	Labs() ([]*db.BasicLabInfo, error)
	// LabTests returns a list of supported single lab tests and test packages
	// for the lab with the provided labID. Returns an ErrorInvalidRequest if
	// labID does not exist.
	LabTests(labID string) (*db.LabTests, error)

	/**** Server Info ****/

	// Faqs returns information about frequently asked questions and help links.
	Faqs() (*db.Faqs, error)
}
