package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiKeyMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		apiKey         string
		headerKey      string
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "Valid API Key",
			apiKey:         "valid-key",
			headerKey:      "Bearer valid-key",
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name:           "Invalid API Key",
			apiKey:         "valid-key",
			headerKey:      "Bearer invalid-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]string{"error": "Invalid api key"},
		},
		{
			name:           "Invalid API Key",
			apiKey:         "valid-key",
			headerKey:      "valid-key",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "The Authorization token should be prefixed with `Bearer`"},
		},
		{
			name:           "No API Key Set",
			apiKey:         "",
			headerKey:      "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]string{"error": "Configure api key to access this endpoint"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := apiKeyMiddleware(tc.apiKey, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
			})

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tc.headerKey != "" {
				req.Header.Set(API_KEY_HEADER_NAME, tc.headerKey)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}

			if rr.Header().Get("Content-Type") != "application/json" {
				t.Error("expected application/json in the response headers")
			}

			if tc.expectedBody != nil {
				var body map[string]string
				err := json.NewDecoder(rr.Body).Decode(&body)
				if err != nil {
					t.Fatalf("Could not decode response body: %v", err)
				}

				if body["error"] != tc.expectedBody["error"] {
					t.Errorf("handler returned unexpected body: got %v want %v", body, tc.expectedBody)
				}
			}
		})
	}
}
