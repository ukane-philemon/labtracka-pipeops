package patient

import (
	"context"
	"net/http"
)

const (
	accessTokenHeader = "X-LabTracka-Access-Token"

	patientEmailCtx = "email"
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

		req.WithContext(context.WithValue(req.Context(), patientEmailCtx, id))
		next.ServeHTTP(res, req)
	})
}
