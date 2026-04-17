package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// decodeJSON decodes the request body into dst.
func decodeJSON(r *http.Request, dst interface{}) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

// ReadBody reads the entire request body and returns the bytes.
func ReadBody(r *http.Request) ([]byte, error) {
	return io.ReadAll(r.Body)
}

// NewBodyReader creates a new ReadCloser from bytes, for restoring r.Body.
func NewBodyReader(data []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(data))
}

