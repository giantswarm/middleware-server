package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Response struct {
	once sync.Once
	w    http.ResponseWriter
}

func (response *Response) Json(result interface{}, code int) error {
	response.once.Do(func() {
		response.w.Header().Add("Content-Type", "application/json")
		response.w.WriteHeader(code)
		if err := json.NewEncoder(response.w).Encode(result); err != nil {
			fmt.Printf("json.NewEncoder: %#v\n", err)
		}
	})

	return nil
}

func (response *Response) Error(content string, code int) error {
	response.once.Do(func() {
		response.w.WriteHeader(code)
		response.w.Write([]byte(content))
	})

	return fmt.Errorf(content)
}

func (response *Response) PlainText(content string, code int) error {
	response.once.Do(func() {
		response.w.WriteHeader(code)
		response.w.Write([]byte(content))
	})

	return nil
}

func (response *Response) NoContent() error {
	response.once.Do(func() {
		response.w.WriteHeader(http.StatusNoContent)
	})

	return nil
}

// Unauthorized sends the http.StatusUnauthorized status code.
// Use this to signal the requestee that the authentication failed.
func (response *Response) Unauthorized(scheme string) error {
	response.once.Do(func() {
		response.w.Header().Add("WWW-Authenticate", scheme)
		response.w.WriteHeader(http.StatusUnauthorized)
	})

	return fmt.Errorf("401 Unauthorized")
}

// Forbidden sends the http.StatusForbidden status.
// Use it to signal that the requestee has no access to the given resource (but auth itself worked).
func (response *Response) Forbidden() error {
	response.once.Do(func() {
		response.w.WriteHeader(http.StatusForbidden)
	})

	return fmt.Errorf("403 Forbidden")
}

func (response *Response) Redirect(location string, code int) error {
	response.once.Do(func() {
		response.w.Header().Set("Location", location)
		response.w.WriteHeader(code)
	})

	return nil
}
