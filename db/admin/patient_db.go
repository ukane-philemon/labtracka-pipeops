package admin

import "github.com/ukane-philemon/labtracka-api/db"

/**** Patient ****/

func (m *MongoDB) PatientLabStats(patientID string) (db.PatientStats, error) {
	return db.PatientStats{}, nil
}

// Results returns all results for the customer with the specified email
// address.
func (m *MongoDB) Results(patientID string) ([]*db.LabResult, error) {
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

/**** Server Info ****/

// Faqs returns information about frequently asked questions and help links.
func (m *MongoDB) Faqs() (*db.Faqs, error) {
	return nil, nil
}
