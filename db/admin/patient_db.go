package admin

import (
	"time"

	"github.com/ukane-philemon/labtracka-api/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**** Patient ****/

func (m *MongoDB) PatientLabStats(patientID string) (db.PatientStats, error) {
	return db.PatientStats{
		TotalNumberOfLabsVisited:     5,
		TotalNumberOfCompletedOrders: 20,
	}, nil
}

// Results returns all results for the patient with the specified email
// address.
func (m *MongoDB) Results(patientID string) ([]*db.LabResult, error) {
	nowUnix := time.Now().Unix()
	return []*db.LabResult{
		{
			ID:                  patientID,
			TestName:            "Malaria",
			LabName:             "Test Lab",
			Status:              db.ResultStatusPending,
			Data:                []string{},
			TurnAroundInSeconds: 0,
			CreatedAt:           uint64(nowUnix),
			LastUpdatedAt:       uint64(nowUnix),
		},
		{
			ID:                  patientID,
			TestName:            "Typhoid",
			LabName:             "Test Lab",
			Status:              db.ResultStatusInProgress,
			Data:                []string{},
			TurnAroundInSeconds: 0,
			CreatedAt:           uint64(nowUnix),
			LastUpdatedAt:       uint64(nowUnix),
		},
		{
			ID:                  patientID,
			TestName:            "Pregnancy",
			LabName:             "Test Lab",
			Status:              db.ResultStatusCompleted,
			Data:                []string{"random base64 file or file url"},
			TurnAroundInSeconds: 60 * 60 * 4,
			CreatedAt:           uint64(nowUnix),
			LastUpdatedAt:       uint64(nowUnix),
		},
	}, nil
}

/**** Labs ****/

// Labs returns a list of available labs.
func (m *MongoDB) Labs() ([]*db.BasicLabInfo, error) {
	return []*db.BasicLabInfo{
		{
			ID:      primitive.NewObjectID().Hex(),
			Name:    "Test Lab",
			LogoURL: "full path to logo url",
			Address: db.Address{
				Coordinates: "",
				HouseNumber: "29",
				StreetName:  "Test street, musa close",
				City:        "Port Harcourt",
				Country:     "Nigeria",
			},
			Featured: true,
		},
		{
			ID:      primitive.NewObjectID().Hex(),
			Name:    "Zion Test Lab",
			LogoURL: "full path to logo url",
			Address: db.Address{
				Coordinates: "",
				HouseNumber: "12",
				StreetName:  "Dynwell street",
				City:        "Port Harcourt",
				Country:     "Nigeria",
			},
			Featured: false,
		},
	}, nil
}

// LabTests returns a list of supported single lab tests and test packages
// for the lab with the provided labID. Returns an ErrorInvalidRequest if
// labID does not exist.
func (m *MongoDB) LabTests(labID string) (*db.LabTests, error) {
	return nil, nil
}

// LabTest returns the information of a lab test from the admin db.
func (m *MongoDB) LabTest(testID string) (*db.LabTest, error) {
	return nil, nil
}

/**** Server Info ****/

// Faqs returns information about frequently asked questions and help links.
func (m *MongoDB) Faqs() (*db.Faqs, error) {
	return nil, nil
}
