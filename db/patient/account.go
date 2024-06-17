package patient

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ukane-philemon/labtracka-api/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// CreateAccount creates a new patient and their information are saved to the
// database. Returns an ErrorInvalidRequest if user email is already tied to
// another patient.
func (m *MongoDB) CreateAccount(req *db.CreateAccountRequest) error {
	if req.Patient == nil || req.DeviceID == "" || req.Password == "" {
		return fmt.Errorf("%w: missing required field(s)", db.ErrorInvalidRequest)
	}

	err := req.Patient.Validate()
	if err != nil {
		return fmt.Errorf("%w: %v", db.ErrorInvalidRequest, err)
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt.GenerateFromPassword error: %w", err)
	}

	patientInfo := &dbPatient{
		PatientInfo:        req.Patient,
		Password:           string(encryptedPassword),
		CreatedAtTimestamp: time.Now().Unix(),
		DeviceID:           req.DeviceID,
	}

	_, err = m.patient.Collection(accountCollection).InsertOne(m.ctx, patientInfo)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: patient already exists", db.ErrorInvalidRequest)
		}
		return fmt.Errorf("accountCollection.InsertOne error: %w", err)
	}

	return nil
}

// PatientID returns the database ID for the patient with the provided email.
func (m *MongoDB) PatientID(email string) (string, error) {
	patientID, err := m.patientID(email)
	if err != nil {
		return "", err
	}
	return patientID, nil
}

// PatientInfo returns the information of the patient with the provided email.
func (m *MongoDB) PatientInfo(email string) (*db.Patient, error) {
	return nil, nil
}

// PatientInfo returns the information of the patient with the provided
// patientID.
func (m *MongoDB) PatientInfoWithID(patientID string) (*db.Patient, error) {
	// Retrieve sub account ids and add to patient info.

	return nil, nil
}

// AddSubAccount adds a new sub account to a patient's profile.
func (m *MongoDB) AddSubAccount(email string, account *db.SubAccount) ([]*db.SubAccountInfo, error) {
	if account == nil {
		return nil, fmt.Errorf("%w: missing required arguments", db.ErrorInvalidRequest)
	}

	patientID, err := m.patientID(email)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(account.Email, email) {
		return nil, fmt.Errorf("%w: cannot use the same email as main account", db.ErrorInvalidRequest)
	}

	// Check if sub account email exists and it's not the main account email.
	subAccountsCollection := m.patient.Collection(subAccountsCollection)
	counts, err := subAccountsCollection.CountDocuments(m.ctx, bson.M{patientIDKey: patientID, mapKey(subAccountsKey, account.Email): account.Email})
	if err != nil {
		return nil, fmt.Errorf("subAccountsCollection.CountDocuments %w", err)
	}

	if counts != 0 {
		return nil, fmt.Errorf("%w: sub account with email %s already exists", db.ErrorInvalidRequest, account.Email)
	}

	// Add sub account to record.
	subAccount := &db.SubAccountInfo{
		ID:         primitive.NewObjectID().Hex(),
		SubAccount: *account,
		Timestamp:  time.Now().Unix(),
	}

	opts := options.Update().SetUpsert(true)
	update := bson.M{setAction: bson.M{mapKey(subAccountsKey, account.Email): subAccount}}
	res, err := subAccountsCollection.UpdateOne(m.ctx, bson.M{patientIDKey: patientID}, update, opts)
	if err != nil {
		return nil, fmt.Errorf("subAccountsCollection.UpdateOne error: %w", err)
	}

	if res.ModifiedCount == 0 {
		return nil, errors.New("patient sub account record was not updated")
	}

	return m.subAccounts(patientID)
}

// SubAccounts returns the sub account for the patient with the provided
// email address.
func (m *MongoDB) SubAccounts(email string) ([]*db.SubAccountInfo, error) {
	patientID, err := m.patientID(email)
	if err != nil {
		return nil, err
	}
	return m.subAccounts(patientID)
}

func (m *MongoDB) subAccounts(patientID string) ([]*db.SubAccountInfo, error) {
	var subAccounts *subAccountRecord
	subAccountsCollection := m.patient.Collection(subAccountsCollection)
	err := subAccountsCollection.FindOne(m.ctx, bson.M{patientIDKey: patientID}).Decode(&subAccounts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []*db.SubAccountInfo{}, nil
		}
		return nil, fmt.Errorf("subAccountsCollection.FindOne error: %w", err)
	}

	var accounts []*db.SubAccountInfo
	for key := range subAccounts.SubAccounts {
		accounts = append(accounts, subAccounts.SubAccounts[key])
	}

	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].Timestamp > accounts[j].Timestamp
	})

	return accounts, nil
}

// RemoveSubAccount removes a sub account from a patient's record and
// returns the remaining sub accounts. Return db.ErrorInvalidRequest if
// subAccountID does not exist.
func (m *MongoDB) RemoveSubAccount(email, subAccountID string) ([]*db.SubAccountInfo, error) {
	return nil, nil
}

// AddNewAddress adds a new address to a patient's profile.
func (m *MongoDB) AddNewAddress(email string, address *db.Address) ([]*db.Address, error) {
	if err := address.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", db.ErrorInvalidRequest, err)
	}

	var patient *dbPatient
	accountsColl := m.patient.Collection(accountCollection)
	err := accountsColl.FindOne(m.ctx, bson.M{emailKey: email}).Decode(&patient)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("account (%s) not found", email)
		}
		return nil, err
	}

	patient.OtherAddress = append(patient.OtherAddress, address)
	_, err = accountsColl.UpdateByID(m.ctx, patient.ID, bson.M{setAction: bson.M{otherAddressKey: patient.OtherAddress}})
	if err != nil {
		return nil, err
	}

	return patient.OtherAddress, nil
}

// SaveProfileImage updates the profile link for a patient.
func (m *MongoDB) SaveProfileImage(email string, profileURL string) error {
	accountsColl := m.patient.Collection(accountCollection)
	res, err := accountsColl.UpdateOne(m.ctx, bson.M{emailKey: email}, bson.M{setAction: bson.M{profileImageKey: profileURL}})
	if err != nil {
		return fmt.Errorf("accountsColl.UpdateOne error: %w", err)
	}

	if res.ModifiedCount == 0 {
		return errors.New("patient profile image was not updated")
	}

	return nil
}
