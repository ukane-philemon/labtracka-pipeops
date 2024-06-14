package patient

import "github.com/ukane-philemon/labtracka-api/db"

// Orders returns all the orders pending orders in the patient db for the
// provided labIDs.
func (m *MongoDB) Orders(labIDs ...string) (map[string]map[string][]*db.Order, error) {
	return nil, nil
}

// AdminStats returns some admin stats for display. If no lab id is returned,
// all current stats will be returned.
func (m *MongoDB) AdminStats(labIDs ...string) (db.AdminStats, error) {
	return db.AdminStats{}, nil
}
