package patient

import "github.com/ukane-philemon/labtracka-api/db"

// Orders returns all the orders pending orders in the patient db for the
// provided labIDs.
func (m *MongoDB) Orders(labIDs ...string) (map[string]map[string][]*db.Order, error) {
	return nil, nil
}
