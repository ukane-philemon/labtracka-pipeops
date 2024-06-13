package db

import (
	"errors"
	"strings"
	"time"

	"github.com/ukane-philemon/labtracka-api/internal/validator"
)

/**** ADMIN TYPES ****/

/**** USER TYPES ****/

// CustomerInfo is information about a customer.
type CustomerInfo struct {
	Name         string             `json:"name"`
	Email        string             `json:"email"`
	PhoneNumber  string             `json:"phone_number" bson:"phone_number"` // optional
	DateOfBirth  time.Time          `json:"date_of_birth" bson:"date_of_birth"`
	Address      *CustomerAddress   `json:"customer_address" bson:"customer_address"`
	OtherAddress []*CustomerAddress `json:"other_address" bson:"other_address"`
	Gender       string             `json:"gender"`
}

func (u *CustomerInfo) Validate() error {
	v := new(validator.Validator)
	v.Check(u.Name != "", "customer name is required")
	v.Check(validator.IsEmail(u.Email), "a valid email address is required")
	v.Check(u.PhoneNumber == "" || validator.IsValidPhoneNumber(u.PhoneNumber), "phone number is invalid")
	v.Check(!u.DateOfBirth.IsZero() && !u.DateOfBirth.After(time.Now()), "please provide a valid date of birth")
	v.Check(time.Since(u.DateOfBirth) > 24*365*18*time.Hour, "you must be at least 18 years or older")
	v.Check(u.Gender == "Male" || u.Gender == "Female", `gender must either be "Male" or "Female"`)
	v.Check(u.Address != nil, "customer address is required")

	if v.HasErrors() {
		return errors.New(strings.Join(v.Errors, ", "))
	}

	if err := u.Address.Validate(); err != nil {
		return err
	}

	for i := range u.OtherAddress {
		if err := u.OtherAddress[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

type CustomerAddress struct {
	Coordinates string `json:"coordinates"`
	HouseNumber string `json:"house_number" bson:"house_number"`
	StreetName  string `json:"street_name" bson:"street_name"`
	City        string `json:"city"`
	Country     string `json:"country"`
}

func (ca *CustomerAddress) Validate() error {
	v := new(validator.Validator)
	v.Check(ca.HouseNumber != "", "house number is required")
	v.Check(ca.StreetName != "", "street name is required")
	v.Check(ca.City != "", "city is required")
	v.Check(ca.Country != "", `country is required`)

	if v.HasErrors() {
		return errors.New(strings.Join(v.Errors, ", "))
	}

	return nil
}

// Customer is the complete information about a customer, including password
// information.
type Customer struct {
	ID              string `json:"id"`
	ProfileImageURL string `json:"profile_image" bson:"profile_image"`
	CustomerInfo
}

// CreateAccountRequest is a struct used to pass around argument for customer
// account creation.
type CreateAccountRequest struct {
	Customer *CustomerInfo
	DeviceID string
	Password string
}

type PatientStats struct {
	TotalNumberOfLabsVisited     int64 `json:"total_number_of_labs_visited"`
	TotalNumberOfCompletedOrders int64 `json:"total_number_of_completed_orders"`
}

type SubAccountInfo struct {
	ID string `json:"id"`
	SubAccount
}

type SubAccount struct {
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Gender       string    `json:"gender"`
	DateOfBirth  time.Time `json:"date_of_birth" bson:"date_of_birth"`
	Relationship string    `json:"relationship"`
	PhoneNumber  string    `json:"phone_number" bson:"phone_number"`
	Address      string    `json:"address"`
}

func (sa *SubAccount) Validate() error {
	v := new(validator.Validator)
	v.Check(sa.Name != "", "customer name is required")
	v.Check(validator.IsEmail(sa.Email), "a valid email address is required")
	v.Check(sa.PhoneNumber == "" || validator.IsValidPhoneNumber(sa.PhoneNumber), "phone number is invalid")
	v.Check(!sa.DateOfBirth.IsZero() && !sa.DateOfBirth.After(time.Now()), "please provide a valid date of birth")
	v.Check(sa.Gender == "Male" || sa.Gender == "Female", `gender must either be "Male" or "Female"`)
	v.Check(sa.Address != "", "customer address is required")

	if v.HasErrors() {
		return errors.New(strings.Join(v.Errors, ", "))
	}

	return nil
}

// LoginRequest is information require by the database implementation to login a
// user (customer or admin).
type LoginRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ClientIP          string `json:"client_ip" bson:"client_ip"`
	DeviceID          string `json:"device_id" bson:"device_id"`
	SaveNewDeviceID   bool   `json:"save_new_device_id" bson:"save_new_device_id"`
	NotificationToken string `json:"notification_token" bson:"notification_token"`
}

// BasicLabInfo is basic information about a laboratory.
type BasicLabInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LogoURL  string `json:"logo_url" bson:"logo_url"`
	Address  string `json:"address"`
	Featured bool   `json:"featured"`
}

// LabTests is information about tests offered by a laboratory.
type LabTests struct {
	Categories   []*TestCategory     `json:"categories"`
	SingleTests  []*LabTest          `json:"single_tests"`
	TestPackages []*LabHealthPackage `json:"test_packages"`
}

type LabTest struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Price                  float64  `json:"price"`
	OldPrice               float64  `json:"old_price" bson:"old_price"`
	Description            string   `json:"description"`
	Gender                 string   `json:"gender"`
	Categories             []string `json:"categories"`
	IsActive               bool     `json:"is_active" bson:"is_active"`
	SampleCollectionMethod []string `json:"sample_collection_method" bson:"sample_collection_method"`
	CreatedAt              uint64   `json:"created_at" bson:"created_at"`
	LastUpdatedAt          uint64   `json:"last_updated_at" bson:"last_updated_at"`
}

type LabHealthPackage struct {
	*LabTest
	Tests []string `json:"tests"` // tests ID when saving to db/test names when retrieving from db.
}

type ResultStatus string

const (
	ResultStatusCompleted  ResultStatus = "Completed"
	ResultStatusPending    ResultStatus = "Pending"
	ResultStatusInProgress ResultStatus = "In Progress"
)

type LabResult struct {
	ID            string       `json:"id"`
	TestName      string       `json:"test_name" bson:"test_name"`
	LabName       string       `json:"lab_name" bson:"lab_name"`
	Status        ResultStatus `json:"status"`
	Data          string       `json:"data"` // base64 encoded or a file url
	CreatedAt     uint64       `json:"created_at" bson:"created_at"`
	LastUpdatedAt uint64       `json:"last_updated_at" bson:"last_updated_at"`
}

type TestCategory struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	IsActive      bool   `json:"is_active" bson:"is_active"`
	CreatedAt     uint64 `json:"created_at" bson:"created_at"`
	LastUpdatedAt uint64 `json:"last_updated_at" bson:"last_updated_at"`
}

type NotificationType string

const (
	NotificationTypePayment = "payment"
	NotificationTypeResults = "result"
	NotificationTypeOrder   = "order"
	NotificationTypeSystem  = "system"
)

type Notification struct {
	ID        string            `json:"id"`
	Type      NotificationType  `json:"type"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Read      bool              `json:"read"`
	Data      map[string]string `json:"data"`
	Timestamp string            `json:"timestamp"`
}

type Faq struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type HelpLink struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type Faqs struct {
	Faqs      []*Faq      `json:"faqs"`
	HelpLinks []*HelpLink `json:"help_links" bson:"help_links"`
}

type OrderTest struct {
	TestID  string  `json:"test_id"` // test ids, can be packages
	Amount  float64 `json:"amount"`
	LabName string  `json:"lab_name"`
}

type Order struct {
	ID           string           `json:"id"`
	Description  string           `json:"description"`
	TotalAmount  float64          `json:"total_amount" bson:"total_amount"`
	Tests        []*OrderTest     `json:"tests"`
	PatientID    string           `json:"patient_id" bson:"patient_id"`
	SubAccountID string           `json:"sub_account_id" bson:"sub_account_id"`
	Status       string           `json:"status"` // Pending/Paid
	Address      *CustomerAddress `json:"patient_address" bson:"patient_address"`
	Timestamp    int64            `json:"timestamp"`
}

type CreateOrderRequest struct {
	Tests          []string         `json:"tests"` // test ids, can be packages
	PatientID      string           `json:"patient_id" bson:"patient_id"`
	SubAccountID   string           `json:"sub_account_id" bson:"sub_account_id"`
	PatientAddress *CustomerAddress `json:"patient_address" bson:"patient_address"`
}

// BankCard holds information about a bank card.
type BankCard struct {
	// ID is the card signature received when the card was bond, a unique ID.
	ID              string     `json:"id"`
	Bin             string     `json:"bin"`
	Last4           string     `json:"last4"`
	ExpirationMonth string     `json:"expirationMonth"`
	ExpirationYear  string     `json:"expirationYear"`
	Type            string     `json:"type"`
	BankName        string     `json:"bankName"`
	AccountName     string     `json:"accountName"`
	Email           string     `json:"email"`
	AuthorizedDate  *time.Time `json:"authorizedDate"`
	Disabled        bool       `json:"disabled"`
	// Error will be empty if this card is valid.
	Error string `json:"error"`
}
