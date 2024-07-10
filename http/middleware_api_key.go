package http

import (
	"encoding/json"
	"net/http"

	"github.com/logdyhq/logdy-core/utils"
)

const API_KEY_HEADER_NAME = "logdy-api-key"

func apiKeyMiddleware(apiKey string, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		key := r.Header.Get(API_KEY_HEADER_NAME)

		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Api key not set in the headers (" + API_KEY_HEADER_NAME + ")",
			})
			return
		}

		if apiKey != "" && key == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Missing '" + API_KEY_HEADER_NAME + "' header with the api key",
			})
			return
		}

		if apiKey != "" && apiKey != key {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid api key",
			})
			return
		}

		utils.Logger.Debugf("API request authorized: %s", r.URL.Path)

		f(w, r)
	}
}
