package admin

import (
	"context"
	"net/http"
)

const (
	accessTokenHeader = "X-LabTracka-Admin-Access-Token"

	adminEmailCtx = "email"
)

// accessTokenValidator checks that the request provides a valid access token.
func (s *Server) accessTokenValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		accessToken := req.Header.Get(accessTokenHeader)
		if accessToken == "" {
			s.authenticationRequired(res, req)
			return
		}

		id, valid := s.jwtManager.IsValidToken(accessToken)
		if !valid {
			s.authenticationRequired(res, req)
			return
		}

		req.WithContext(context.WithValue(req.Context(), adminEmailCtx, id))
		next.ServeHTTP(res, req)
	})
}

func (s *Server) reqAuthID(req *http.Request) string {
	emailCtxValue := req.Context().Value(adminEmailCtx)
	if emailCtxValue != nil {
		return emailCtxValue.(string)
	}
	return ""
}
