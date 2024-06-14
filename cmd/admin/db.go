package admin

import "github.com/ukane-philemon/labtracka-api/db"

type Database interface {
	Shutdown()
}

type PatientDatabase interface {
	// Orders returns all the orders pending orders in the patient db for the
	// provided labIDs.
	Orders(labIDs ...string) (map[string]map[string][]*db.Order, error)
}
