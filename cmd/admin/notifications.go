package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ukane-philemon/labtracka-api/db"
)

// handleMarkNotificationAsRead handles the "POST /mark-notifications-as-read"
// endpoint and marks the notifications provided as query params noteID,
// multiple IDs must be supported by a coma ",".
func (s *Server) handleMarkNotificationAsRead(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.badRequest(res, req, "you not authorized to access this resource")
		return
	}

	noteIDs := strings.Split(req.URL.Query().Get("noteID"), ",")
	if len(noteIDs) == 0 {
		s.badRequest(res, req, "missing notedID")
		return
	}

	err := s.db.MarkNotificationsAsRead(authID, noteIDs...)
	if err != nil {
		if errors.Is(err, db.ErrorInvalidRequest) {
			s.badRequest(res, req, trimErrorInvalidRequest(err))
		} else {
			s.serverError(res, req, fmt.Errorf("db.MarkNotificationsAsRead error: %w", err))
		}
		return
	}

	s.sendSuccessResponse(res, req, "Notification(s) marked as read")
}

func trimErrorInvalidRequest(err error) string {
	return strings.TrimPrefix(err.Error(), db.ErrorInvalidRequest.Error()+": ")
}
