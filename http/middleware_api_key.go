package http

import (
	"net/http"
	"strings"

	"github.com/logdyhq/logdy-core/utils"
)

const API_KEY_HEADER_NAME = "Authorization"

func apiKeyMiddleware(apiKey string, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiKey == "" {
			httpError("Configure api key to access this endpoint", w, http.StatusUnauthorized)
			return
		}

		key := r.Header.Get(API_KEY_HEADER_NAME)

		if !strings.HasPrefix(key, "Bearer ") {
			httpError("The Authorization token should be prefixed with `Bearer`", w, http.StatusBadRequest)
			return
		}

		if key == "" {
			httpError("Missing '"+API_KEY_HEADER_NAME+"' header with the api key", w, http.StatusUnauthorized)
			return
		}

		key, _ = strings.CutPrefix(key, "Bearer ")
		if apiKey != key {
			httpError("Invalid api key", w, http.StatusUnauthorized)
			return
		}

		utils.Logger.Debugf("API request authorized: %s", r.URL.Path)

		f(w, r)
	}
}
