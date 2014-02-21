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
