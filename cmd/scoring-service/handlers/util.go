package handlers

import (
	"encoding/json"
	"net/http"
)

// decodeJSON decodes the request body into dst.
// Returns an error if the body is not valid JSON.
func decodeJSON(r *http.Request, dst interface{}) error {
	return json.NewDecoder(r.Body).Decode(dst)
}
