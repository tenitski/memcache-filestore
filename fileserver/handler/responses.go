package handler

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type errorResponse struct {
	Error string `json:"error"`
}

func respondWithData(w http.ResponseWriter, r *http.Request, code int, body []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(code)

	_, err := w.Write(body)
	if err != nil {
		log.WithError(err).Error("Error while sending response")
	}
}

func respondWithStatusCode(w http.ResponseWriter, r *http.Request, code int) {
	w.WriteHeader(code)
}

func respondWithError(w http.ResponseWriter, r *http.Request, code int, responseErr error) {
	// Build json error response
	response := errorResponse{
		Error: responseErr.Error(),
	}

	body, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Send it
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err = w.Write(body)
	if err != nil {
		log.WithError(err).Error("Error while sending response")
	}
}
