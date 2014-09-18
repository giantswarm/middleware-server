package server

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	w http.ResponseWriter
}

func (response *Response) Json(result interface{}, code int) error {
	response.w.Header().Add("Content-Type", "application/json")
	response.w.WriteHeader(code)
	return json.NewEncoder(response.w).Encode(result)
}

func (response *Response) Error(message string, code int) error {
	return response.PlainText(message, code)
}

func (response *Response) PlainText(content string, code int) error {
	response.w.WriteHeader(code)
	response.w.Write([]byte(content))
	return nil
}

func (response *Response) NoContent() error {
	response.w.WriteHeader(http.StatusNoContent)
	return nil
}

// Unauthorized sends the http.StatusUnauthorized status code.
// Use this to signal the requestee that the authentication failed.
func (response *Response) Unauthorized(scheme string) error {
	response.w.Header().Add("WWW-Authenticate", scheme)
	response.w.WriteHeader(http.StatusUnauthorized)
	return nil
}

// Forbidden sends the http.StatusForbidden status.
// Use it to signal that the requestee has no access to the given resource (but auth itself worked).
func (response *Response) Forbidden() error {
	response.w.WriteHeader(http.StatusForbidden)
}

func (response *Response) Redirect(location string, code int) error {
	response.w.Header().Set("Location", location)
	response.w.WriteHeader(code)
	return nil
}
