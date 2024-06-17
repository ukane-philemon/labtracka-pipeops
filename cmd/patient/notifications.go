package patient

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ukane-philemon/labtracka-api/db"
)

// handleGetNotifications handles the "GET /notifications" endpoint and returns
// all notification for the logged in patient.
func (s *Server) handleGetNotifications(res http.ResponseWriter, req *http.Request) {
	authID := s.reqAuthID(req)
	if authID == "" {
		s.badRequest(res, req, "you not authorized to access this resource")
		return
	}

	notifications, err := s.db.Notifications(authID)
	if err != nil {
		s.serverError(res, req, fmt.Errorf("db.Notifications error: %w", err))
		return
	}

	s.sendSuccessResponseWithData(res, req, notifications)
}

// handleMarkNotificationAsRead handles the "PATCH /mark-notifications-as-read"
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
