package db

import (
	"errors"
	"strings"
	"time"

	"github.com/ukane-philemon/labtracka-api/internal/validator"
)

/**** ADMIN TYPES ****/

/**** USER TYPES ****/

// PatientInfo is information about a patient.
type PatientInfo struct {
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PhoneNumber  string     `json:"phone_number" bson:"phone_number"` // optional
	DateOfBirth  time.Time  `json:"date_of_birth" bson:"date_of_birth"`
	Address      *Address   `json:"patient_address" bson:"patient_address"`
	OtherAddress []*Address `json:"other_address" bson:"other_address"`
	Gender       string     `json:"gender"`
}

func (p *PatientInfo) Validate() error {
	v := new(validator.Validator)
	v.Check(p.Name != "", "patient name is required")
	v.Check(validator.IsEmail(p.Email), "a valid email address is required")
	v.Check(p.PhoneNumber == "" || validator.IsValidPhoneNumber(p.PhoneNumber), "phone number is invalid")
	v.Check(!p.DateOfBirth.IsZero() && !p.DateOfBirth.After(time.Now()), "please provide a valid date of birth")
	v.Check(time.Since(p.DateOfBirth) > 24*365*18*time.Hour, "you must be at least 18 years or older")
	v.Check(p.Gender == "Male" || p.Gender == "Female", `gender must either be "Male" or "Female"`)
	v.Check(p.Address != nil, "patient address is required")

	if v.HasErrors() {
		return errors.New(strings.Join(v.Errors, ", "))
	}

	if err := p.Address.Validate(); err != nil {
		return err
	}

	for i := range p.OtherAddress {
		if err := p.OtherAddress[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

type Address struct {
	Coordinates string `json:"coordinates"`
	HouseNumber string `json:"house_number" bson:"house_number"`
	StreetName  string `json:"street_name" bson:"street_name"`
	City        string `json:"city"`
	Country     string `json:"country"`
}

func (a *Address) Validate() error {
	v := new(validator.Validator)
	v.Check(a.HouseNumber != "", "house number is required")
	v.Check(a.StreetName != "", "street name is required")
	v.Check(a.City != "", "city is required")
	v.Check(a.Country != "", `country is required`)

	if v.HasErrors() {
		return errors.New(strings.Join(v.Errors, ", "))
	}

	return nil
}

// Patient is the complete information about a patient, including password
// information.
type Patient struct {
	ID              string   `json:"id"`
	ProfileImageURL string   `json:"profile_image" bson:"profile_image"`
	SubAccounts     []string `json:"sub_accounts" bson:"sub_accounts"`
	PatientInfo     `bson:"inline"`
}

func (p *Patient) HasSubAccount(subAccountID string) bool {
	for _, subAccID := range p.SubAccounts {
		if subAccID == subAccID {
			return true
		}
	}
	return false
}

// CreateAccountRequest is a struct used to pass around argument for patient's
// account creation.
type CreateAccountRequest struct {
	Patient  *PatientInfo
	DeviceID string
	Password string
}

type PatientStats struct {
	TotalNumberOfLabsVisited     int64 `json:"total_number_of_labs_visited"`
	TotalNumberOfCompletedOrders int64 `json:"total_number_of_completed_orders"`
}

type AdminStats struct {
	TotalUsers         int64 `json:"total_users"`
	TotalRevenue       int64 `json:"total_revenue"`
	TotalTestOrders    int64 `json:"total_test_orders"`
	TotalPendingOrders int64 `json:"total_pending_orders"`
}

type SubAccountInfo struct {
	ID         string `json:"id"`
	SubAccount `bson:"inline"`
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
	v.Check(sa.Name != "", "patient name is required")
	v.Check(validator.IsEmail(sa.Email), "a valid email address is required")
	v.Check(sa.PhoneNumber == "" || validator.IsValidPhoneNumber(sa.PhoneNumber), "phone number is invalid")
	v.Check(!sa.DateOfBirth.IsZero() && !sa.DateOfBirth.After(time.Now()), "please provide a valid date of birth")
	v.Check(sa.Gender == "Male" || sa.Gender == "Female", `gender must either be "Male" or "Female"`)
	v.Check(sa.Address != "", "patient address is required")

	if v.HasErrors() {
		return errors.New(strings.Join(v.Errors, ", "))
	}

	return nil
}

// LoginRequest is information require by the database implementation to login a
// user (patient or admin).
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
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	LogoURL  string  `json:"logo_url" bson:"logo_url"`
	Address  Address `json:"address"`
	Featured bool    `json:"featured"`
}

type AdminLabInfo struct {
	BasicLabInfo
	CreatedAt     int64 `json:"created_at"`
	LastUpdatedAt int64 `json:"last_updated_at"`
}

// LabTests is information about tests offered by a laboratory.
type LabTests struct {
	Categories []*TestCategory `json:"categories"`
	Tests      []*LabTest      `json:"single_tests"`
}

type LabTest struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	LabID                  string   `json:"lab_id"`
	LabName                string   `json:"lab_name"`
	Price                  float64  `json:"price"`
	OldPrice               float64  `json:"old_price" bson:"old_price"`
	Description            string   `json:"description"`
	Gender                 string   `json:"gender"`
	Categories             []string `json:"categories"`
	IsDisabled             bool     `json:"is_disabled" bson:"is_disabled"`
	SampleCollectionMethod []string `json:"sample_collection_method" bson:"sample_collection_method"`
	// Tests is an array of tests ID when saving to db/test names when
	// retrieving from db. This will be non-empty for test packages.
	Tests         []string `json:"tests"`
	CreatedAt     int64    `json:"created_at" bson:"created_at"`
	LastUpdatedAt int64    `json:"last_updated_at" bson:"last_updated_at"`
}

type ResultStatus string

const (
	ResultStatusCompleted  ResultStatus = "Completed"
	ResultStatusPending    ResultStatus = "Pending"
	ResultStatusInProgress ResultStatus = "In Progress"
)

type LabResult struct {
	ID                  string       `json:"id"` // order id
	TestName            string       `json:"test_name" bson:"test_name"`
	LabName             string       `json:"lab_name" bson:"lab_name"`
	Status              ResultStatus `json:"status"`
	Data                []string     `json:"data"` // base64 encoded file or a file url
	TurnAroundInSeconds int64        `json:"turn_around_in_seconds"`
	CreatedAt           int64        `json:"created_at" bson:"created_at"`
	LastUpdatedAt       int64        `json:"last_updated_at" bson:"last_updated_at"`
}

type TestCategory struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	IsActive      bool   `json:"is_active" bson:"is_active"`
	CreatedAt     int64  `json:"created_at" bson:"created_at"`
	LastUpdatedAt int64  `json:"last_updated_at" bson:"last_updated_at"`
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
	Title   string `json:"title"`
	Link    string `json:"link"`
	IsVideo bool   `json:"is_video"`
}

type Faqs struct {
	Faqs      []*Faq      `json:"faqs"`
	HelpLinks []*HelpLink `json:"help_links" bson:"help_links"`
}

type OrderTest struct {
	LabID    string  `json:"lab_id"`
	LabName  string  `json:"lab_name"`
	TestID   string  `json:"test_id"` // test ids, can be packages
	TestName string  `json:"test_name"`
	Amount   float64 `json:"amount"`
}

type Order struct {
	ID string `json:"id"`
	OrderInfo
	Timestamp int64 `json:"timestamp"`
}

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "Pending Payment"
	OrderStatusPaid           OrderStatus = "Paid"
)

type OrderInfo struct {
	Description  string       `json:"description"`
	TotalAmount  float64      `json:"total_amount" bson:"total_amount"`
	Fee          float64      `json:"fee"`
	Tests        []*OrderTest `json:"tests"`
	PatientID    string       `json:"patient_id" bson:"patient_id"`
	SubAccountID string       `json:"sub_account_id" bson:"sub_account_id"`
	Status       OrderStatus  `json:"status"` // Pending/Paid
	Address      *Address     `json:"patient_address" bson:"patient_address"`
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

type Admin struct {
	ID              string `json:"id"`
	ProfileImageURL string `json:"profile_image" bson:"profile_image"`
	AdminInfo
	SuperAdmin  bool     `json:"super_admin"`
	ServerAdmin bool     `json:"server_admin"`
	LabIDs      []string `json:"lab_ids"`
}

type AdminInfo struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number" bson:"phone_number"` // optional
}
