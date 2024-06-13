package patient

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/ukane-philemon/labtracka-api/internal/response"
	"github.com/ukane-philemon/labtracka-api/internal/validator"
)

func (s *Server) reportServerError(r *http.Request, err error) {
	var (
		message = err.Error()
		method  = r.Method
		url     = r.URL.String()
		trace   = string(debug.Stack())
	)

	requestAttrs := slog.Group("request", "method", method, "url", url)
	s.logger.Error(message, requestAttrs, "trace", trace)

	if s.mailer != nil && s.cfg.ServerEmail != "" {
		data := map[string]any{
			"BaseURL": s.baseURL,
		}

		data["Message"] = message
		data["RequestMethod"] = method
		data["RequestURL"] = url
		data["Trace"] = trace

		err := s.mailer.Send(s.cfg.ServerEmail, data, "error-notification.tmpl")
		if err != nil {
			trace = string(debug.Stack())
			s.logger.Error(err.Error(), requestAttrs, "trace", trace)
		}
	}
}

func (s *Server) errorMessage(w http.ResponseWriter, r *http.Request, status int, message string) {
	s.errorMessageWithHeader(w, r, status, message, nil)
}

func (s *Server) errorMessageWithHeader(w http.ResponseWriter, r *http.Request, status int, message string, headers http.Header) {
	message = strings.ToUpper(message[:1]) + message[1:]

	err := response.JSONWithHeaders(w, status, map[string]string{"error": message}, headers)
	if err != nil {
		s.reportServerError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.reportServerError(r, err)

	message := "The server encountered a problem and could not process your request"
	s.errorMessage(w, r, http.StatusInternalServerError, message)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	s.errorMessage(w, r, http.StatusNotFound, message)
}

func (s *Server) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	s.errorMessage(w, r, http.StatusMethodNotAllowed, message)
}

func (s *Server) badRequest(w http.ResponseWriter, r *http.Request, errMessage string) {
	s.errorMessage(w, r, http.StatusBadRequest, errMessage)
}

func (s *Server) failedValidation(w http.ResponseWriter, r *http.Request, v validator.Validator) {
	err := response.JSON(w, http.StatusUnprocessableEntity, v)
	if err != nil {
		s.serverError(w, r, err)
	}
}

func (s *Server) authenticationRequired(w http.ResponseWriter, r *http.Request) {
	s.errorMessage(w, r, http.StatusUnauthorized, "You must be authenticated to access this resource")
}

func (s *Server) basicAuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	message := "You must be authenticated to access this resource"
	s.errorMessageWithHeader(w, r, http.StatusUnauthorized, message, headers)
}
