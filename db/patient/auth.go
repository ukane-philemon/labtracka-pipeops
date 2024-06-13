package patient

import (
	"errors"
	"fmt"
	"time"

	"github.com/ukane-philemon/labtracka-api/db"
	"github.com/ukane-philemon/labtracka-api/internal/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// LoginCustomer logs a customer into their account. Returns an
// ErrorInvalidRequest is user email or password is invalid/not correct or does
// not exist or an ErrorOTPRequired if otp validation is required for this
// account.
func (m *MongoDB) LoginCustomer(loginReq *db.LoginRequest) (*db.Customer, error) {
	if validator.AnyValueEmpty(loginReq.Email, loginReq.Password, loginReq.ClientIP, loginReq.DeviceID) {
		return nil, errors.New("missing required field(s)")
	}

	var customer *dbCustomer
	accountsColl := m.customer.Collection(accountCollection)
	err := accountsColl.FindOne(m.ctx, bson.M{mapKey(customerInfoKey, emailKey): loginReq.Email}).Decode(&customer)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: incorrect email or password", db.ErrorInvalidRequest)
		}
		return nil, err
	}

	// Check password.
	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(loginReq.Password)); err != nil {
		return nil, fmt.Errorf("%w: incorrect email or password", db.ErrorInvalidRequest)
	}

	// Check if this is a new deviceID and ensure its validated.
	session, err := m.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("m.client.StartSession error: %w", err)
	}
	defer session.EndSession(m.ctx)

	if loginReq.DeviceID != customer.DeviceID {
		if loginReq.SaveNewDeviceID {
			_, err = session.WithTransaction(m.ctx, func(ctx mongo.SessionContext) (interface{}, error) {
				return accountsColl.UpdateByID(ctx, customer.ID, bson.M{"$set": bson.M{deviceIDKey: loginReq.DeviceID}})
			})
		} else {
			err = fmt.Errorf("%w: otp validation is required for new device", db.ErrorOTPRequired)
		}
	}
	if err != nil {
		return nil, err
	}

	// Log the login ip address.
	_, err = session.WithTransaction(m.ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		filter := bson.M{dbIDKey: customer.ID}
		update := bson.M{"$set": bson.M{mapKey(lastLoginKey, loginReq.ClientIP): time.Now().Unix()}}
		opts := options.FindOneAndUpdate().SetUpsert(true)
		res := m.customer.Collection(loginRecordCollection).FindOneAndUpdate(ctx, filter, update, opts)
		return res, res.Err()
	})

	err = session.CommitTransaction(m.ctx)
	if err != nil {
		return nil, fmt.Errorf("error committing session: %w", err)
	}

	return &db.Customer{
		ID:              customer.ID.String(),
		CustomerInfo:    *customer.CustomerInfo,
		ProfileImageURL: customer.ProfileImage,
	}, nil
}

// ResetPassword reset the password of an existing customer. Returns an
// ErrorInvalidRequest if the email is not tied to an existing customer.
func (m *MongoDB) ResetPassword(email, password string) error {
	return nil
}

// ChangePassword updates the password for an existing customer. Returns an
// ErrorInvalidRequest if email is not tied to an existing customer or
// current password is incorrect.
func (m *MongoDB) ChangePassword(email, currentPassword, newPassword string) error {
	return nil
}

// Notifications returns all the notifications for customer sorted by unread
// first.
func (m *MongoDB) Notifications(email string) ([]*db.Notification, error) {
	return nil, nil
}

// MarkNotificationsAsRead marks the notifications with the provided noteIDs
// as read.
func (m *MongoDB) MarkNotificationsAsRead(email string, noteIDs ...string) error {
	return nil
}

// Results returns all results for the customer with the specified email
// address.
func (m *MongoDB) Results(email string) ([]*db.LabResult, error) {
	return nil, nil
}

// Faqs returns information about frequently asked questions and help links.
func (m *MongoDB) Faqs() (*db.Faqs, error) {
	return nil, nil
}

/**** Labs ****/

// Labs returns a list of available labs.
func (m *MongoDB) Labs() ([]*db.BasicLabInfo, error) {
	return nil, nil
}

// LabTests returns a list of supported single lab tests and test packages
// for the lab with the provided labID. Returns an ErrorInvalidRequest if
// labID does not exist.
func (m *MongoDB) LabTests(labID string) (*db.LabTests, error) {
	return nil, nil
}
