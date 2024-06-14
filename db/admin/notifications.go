package admin

import "github.com/ukane-philemon/labtracka-api/db"

// Notifications returns all the notifications for admin sorted by unread
// first.
func (m *MongoDB) Notifications(email string) ([]*db.Notification, error) {
	return nil, nil
}

// MarkNotificationsAsRead marks the notifications with the provided noteIDs
// as read.
func (m *MongoDB) MarkNotificationsAsRead(email string, noteIDs ...string) error {
	return nil
}
