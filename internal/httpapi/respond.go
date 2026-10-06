package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func jsonResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func fail(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}
func decode(r *http.Request, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, (2<<20)+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("request must contain one JSON value")
	}
	return nil
}
