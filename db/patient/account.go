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
