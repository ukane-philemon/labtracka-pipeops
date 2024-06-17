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

// LoginPatient logs a patient into their account. Returns an
// ErrorInvalidRequest is user email or password is invalid/not correct or does
// not exist or an ErrorOTPRequired if otp validation is required for this
// account.
func (m *MongoDB) LoginPatient(loginReq *db.LoginRequest) (*db.Patient, error) {
	if validator.AnyValueEmpty(loginReq.Email, loginReq.Password, loginReq.ClientIP, loginReq.DeviceID) {
		return nil, errors.New("missing required field(s)")
	}

	var patient *dbPatient
	accountsColl := m.patient.Collection(accountCollection)
	err := accountsColl.FindOne(m.ctx, bson.M{emailKey: loginReq.Email}).Decode(&patient)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: incorrect email or password", db.ErrorInvalidRequest)
		}
		return nil, err
	}

	// Check password.
	if err := bcrypt.CompareHashAndPassword([]byte(patient.Password), []byte(loginReq.Password)); err != nil {
		return nil, fmt.Errorf("%w: incorrect email or password", db.ErrorInvalidRequest)
	}

	// Check if this is a new deviceID and ensure its validated.
	session, err := m.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("m.client.StartSession error: %w", err)
	}
	defer session.EndSession(m.ctx)

	if loginReq.DeviceID != patient.DeviceID {
		if loginReq.SaveNewDeviceID {
			_, err = session.WithTransaction(m.ctx, func(ctx mongo.SessionContext) (interface{}, error) {
				return accountsColl.UpdateByID(ctx, patient.ID, bson.M{setAction: bson.M{deviceIDKey: loginReq.DeviceID}})
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
		filter := bson.M{dbIDKey: patient.ID}
		update := bson.M{setAction: bson.M{mapKey(lastLoginKey, loginReq.ClientIP): time.Now().Unix()}}
		opts := options.FindOneAndUpdate().SetUpsert(true)
		res := m.patient.Collection(loginRecordCollection).FindOneAndUpdate(ctx, filter, update, opts)
		return res, res.Err()
	})

	err = session.CommitTransaction(m.ctx)
	if err != nil {
		return nil, fmt.Errorf("error committing session: %w", err)
	}

	return &db.Patient{
		ID:              patient.ID.Hex(),
		PatientInfo:     *patient.PatientInfo,
		ProfileImageURL: patient.ProfileImage,
	}, nil
}

// ResetPassword reset the password of an existing patient. Returns an
// ErrorInvalidRequest if the email is not tied to an existing patient.
func (m *MongoDB) ResetPassword(email, password string) error {
	var patient *dbPatient
	accountsColl := m.patient.Collection(accountCollection)
	opts := options.FindOne().SetProjection(bson.M{dbIDKey: 1, passwordKey: 1})
	err := accountsColl.FindOne(m.ctx, bson.M{emailKey: email}, opts).Decode(&patient)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("%w: incorrect email or password", db.ErrorInvalidRequest)
		}
		return err
	}

	// Check for password reuse.
	if err := bcrypt.CompareHashAndPassword([]byte(patient.Password), []byte(password)); err == nil {
		return fmt.Errorf("%w: cannot reuse password", db.ErrorInvalidRequest)
	}

	res, err := accountsColl.UpdateOne(m.ctx, patient.ID, bson.M{passwordKey: password}, options.Update().SetUpsert(false))
	if err != nil {
		return fmt.Errorf("accountsColl.UpdateOne error: %w", err)
	}

	if res.ModifiedCount == 0 {
		return errors.New("patient password was not reset")
	}

	return nil
}

// ChangePassword updates the password for an existing patient. Returns an
// ErrorInvalidRequest if email is not tied to an existing patient or
// current password is incorrect.
func (m *MongoDB) ChangePassword(email, currentPassword, newPassword string) error {
	return nil
}

// Notifications returns all the notifications for patient sorted by unread
// first.
func (m *MongoDB) Notifications(email string) ([]*db.Notification, error) {
	patientID, err := m.patientID(email)
	if err != nil {
		return nil, err
	}

	notificationsCollection := m.patient.Collection(notificationsCollection)
	cur, err := notificationsCollection.Find(m.ctx, bson.M{patientIDKey: patientID})
	if err != nil {
		return nil, err
	}

	var patientNotes []*db.Notification
	err = cur.Decode(&patientNotes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode patient notifications: %w", err)
	}

	return patientNotes, nil
}

// MarkNotificationsAsRead marks the notifications with the provided noteIDs
// as read.
func (m *MongoDB) MarkNotificationsAsRead(email string, noteIDs ...string) error {
	patientID, err := m.patientID(email)
	if err != nil {
		return err
	}

	notificationsCollection := m.patient.Collection(notificationsCollection)
	for _, noteID := range noteIDs {
		filter := bson.M{patientIDKey: patientID, idKey: noteID}
		res, err := notificationsCollection.UpdateOne(m.ctx, filter, bson.M{setAction: bson.M{readKey: true}}, options.Update().SetUpsert(false))
		if err != nil {
			return err
		}

		if res.ModifiedCount == 0 {
			return fmt.Errorf("%w: note with id %s does not exist", db.ErrorInvalidRequest, noteID)
		}
	}

	return nil
}
