package patient

import (
	"fmt"
	"time"

	"github.com/ukane-philemon/labtracka-api/db"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var passwordEncryptionCost = 12

// CreateAccount creates a new customer and their information are saved to the
// database. Returns an ErrorInvalidRequest if user email is already tied to
// another customer.
func (m *MongoDB) CreateAccount(req *db.CreateAccountRequest) error {
	if req.Customer == nil || req.DeviceID == "" || req.Password == "" {
		return fmt.Errorf("%w: missing required field(s)", db.ErrorInvalidRequest)
	}

	err := req.Customer.Validate()
	if err != nil {
		return fmt.Errorf("%w: %v", db.ErrorInvalidRequest, err)
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), passwordEncryptionCost)
	if err != nil {
		return fmt.Errorf("bcrypt.GenerateFromPassword error: %w", err)
	}

	customerInfo := &dbCustomer{
		CustomerInfo:       req.Customer,
		Password:           string(encryptedPassword),
		CreatedAtTimestamp: time.Now().Unix(),
		DeviceID:           req.DeviceID,
	}

	_, err = m.customer.Collection(accountCollection).InsertOne(m.ctx, customerInfo)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: customer already exists", db.ErrorInvalidRequest)
		}
		return fmt.Errorf("accountCollection.InsertOne error: %w", err)
	}

	return nil
}

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
