package admin

import "github.com/ukane-philemon/labtracka-api/db"

// UpdateFaqs updates the faqs in the database. This is a super admin only
// feature.
func (m *MongoDB) UpdateFaqs(faq *db.Faqs) (*db.Faqs, error) {
	return nil, nil
}
