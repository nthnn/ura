package util

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/nthnn/ura/logger"
)

func failedWriteFallback(w http.ResponseWriter) {
	_, err := w.Write([]byte("{\"status\": \"error\", \"message\": \"Internal error occurred\"}"))
	if err != nil {
		logger.Error("Error writing failed fallback response: %s", err.Error())
	}
}

func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		logger.Error("Error encoding JSON: %s", err.Error())
		failedWriteFallback(w)

		return
	}

	w.Write(buf.Bytes())
}

func WriteJSONError(w http.ResponseWriter, message string) {
	data := map[string]string{
		"status":  "error",
		"message": message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		logger.Error("Error encoding JSON: %s", err.Error())
		failedWriteFallback(w)

		return
	}

	w.Write(buf.Bytes())
}
