package request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	return decodeJSON(w, r, dst, false)
}

func DecodeJSONStrict(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	return decodeJSON(w, r, dst, true)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}, disallowUnknownFields bool) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	if disallowUnknownFields {
		dec.DisallowUnknownFields()
	}

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// CallAPI sets the auth header information as "Bearer" if provided and makes a
// call to the provided endpoint and returns the response code and bytes.
func CallAPI(ctx context.Context, method, url string, authToken string, requestBody interface{}, customHeaders ...map[string]string) (int, []byte, error) {
	if authToken != "" && !strings.HasPrefix(authToken, "Bearer ") {
		authToken = fmt.Sprintf("Bearer %s", authToken)
	}
	return callAPI(ctx, method, url, authToken, requestBody, customHeaders...)
}

// callAPI sets the auth header information and makes a call to the provided
// endpoint and returns the response code and bytes.
func callAPI(ctx context.Context, method, url string, authorization string, requestBody interface{}, customHeaders ...map[string]string) (int, []byte, error) {
	var reqBody io.Reader
	if method == "POST" {
		body, err := json.Marshal(requestBody)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(body)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return 0, nil, fmt.Errorf("new request: %w", err)
	}

	if authorization != "" {
		httpReq.Header.Add("Authorization", authorization)
	}

	if reqBody != nil {
		httpReq.Header.Add("Content-Type", "application/json")
	}

	for _, headers := range customHeaders {
		for key, value := range headers {
			httpReq.Header.Add(key, value)
		}
	}

	reply, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return 0, nil, fmt.Errorf("%s %s: %w", method, httpReq.URL.String(), err)
	}
	defer reply.Body.Close()

	respBody, err := io.ReadAll(reply.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("read response body: %w", err)
	}

	return reply.StatusCode, respBody, nil
}
