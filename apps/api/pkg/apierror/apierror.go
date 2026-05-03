package apierror

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

func New(status int, code, message string) Error {
	return Error{Status: status, Code: code, Message: message}
}

func BadRequest(message string) Error {
	return New(http.StatusBadRequest, "bad_request", message)
}

func Unauthorized(message string) Error {
	return New(http.StatusUnauthorized, "unauthorized", message)
}

func Forbidden(message string) Error {
	return New(http.StatusForbidden, "forbidden", message)
}

func NotFound(message string) Error {
	return New(http.StatusNotFound, "not_found", message)
}

func ServiceUnavailable(message string) Error {
	return New(http.StatusServiceUnavailable, "service_unavailable", message)
}

func Internal(message string) Error {
	return New(http.StatusInternalServerError, "internal_error", message)
}

func Write(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr Error
	if !errors.As(err, &apiErr) {
		apiErr = Internal("Something went wrong.")
	}
	if apiErr.Status == 0 {
		apiErr.Status = http.StatusInternalServerError
	}

	Respond(w, r, apiErr.Status, map[string]any{"error": apiErr})
}

func Respond(w http.ResponseWriter, _ *http.Request, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(payload)
}
