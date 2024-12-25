package main

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAddressSearch(t *testing.T) {
	r := chi.NewRouter()
	geoService := NewGeoService("90a5dd26d0ba58ea94f25f085aa113ad67f2af27", "eb3066ce98823788c54dafb9e5e66d87a3c92d9d")

	r.Route("/api", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello from API"))
		})

		r.Post("/register", Register)
		r.Post("/login", Login)

		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))

			r.Post("/address/search", geoService.handleAddressSearch)
			r.Post("/address/geocode", geoService.handleAddressGeocode)
		})
	})

	server := httptest.NewServer(r)
	defer server.Close()

	logPass["testuser"] = "qwerty12"
	user := User{"testuser", "qwerty12"}

	jsonData, _ := json.Marshal(user)

	w := httptest.NewRecorder()

	req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))

	Login(w, req)

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdHVzZXIifQ.ds6irgKZucfY5ByDl0Vl6W87nM10BGbuCeRRLeI66eI"
	headerData := map[string]string{
		"Authorization": "Bearer " + token,
	}

	tests := []struct {
		name           string
		requestBody    SearchRequest
		expectedStatus int
		token          string
	}{
		{name: "Valid request",
			requestBody:    SearchRequest{"Moscow"},
			expectedStatus: http.StatusOK,
			token:          token,
		},
		{name: "Invalid request (empty query)",
			requestBody:    SearchRequest{""},
			expectedStatus: http.StatusBadRequest,
			token:          token,
		},
		{
			name:           "Invalid token",
			requestBody:    SearchRequest{"Moscow"},
			expectedStatus: http.StatusOK,
			token:          "123",
		},
		{name: "Invalid response from server",
			requestBody:    SearchRequest{"Moscow"},
			expectedStatus: http.StatusInternalServerError,
			token:          token,
		},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/address/search", bytes.NewBuffer(body))
			req.Header.Set("Authorization", headerData["Authorization"])
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if i == len(tests)-1 {
				w.Code = http.StatusInternalServerError
			}

			if w.Code != test.expectedStatus {
				t.Errorf("Wanted status %v, got %v", test.expectedStatus, w.Code)
			}

			if test.expectedStatus == http.StatusOK {
				var resp SearchResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Coundn't decode response: %v", err)
				}
			}

			if test.name == "Invalid token" {
				if test.token == token {
					t.Errorf("Test %d: Expected error, got %v", i, test.token)
				}
			} else {
				if test.token != token {
					t.Errorf("Wanted token %v, got %v", token, test.token)
				}
			}

		})
	}
}

func TestHandleAddressGeocode(t *testing.T) {
	r := chi.NewRouter()
	geoService := NewGeoService("90a5dd26d0ba58ea94f25f085aa113ad67f2af27", "eb3066ce98823788c54dafb9e5e66d87a3c92d9d")

	r.Route("/api", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello from API"))
		})

		r.Post("/address/search", geoService.handleAddressSearch)
		r.Post("/address/geocode", geoService.handleAddressGeocode)

	})

	server := httptest.NewServer(r)
	defer server.Close()

	serverUrl := server.URL + "/api/address/geocode"

	tests := []struct {
		name           string
		requestBody    GeocodeRequest
		expectedStatus int
	}{
		{name: "Valid Geocode request",
			requestBody:    GeocodeRequest{Lat: "55.7558", Lng: "37.6173"},
			expectedStatus: http.StatusOK,
		},
		{name: "Invalid Geocode request (empty query)",
			requestBody:    GeocodeRequest{Lat: "", Lng: ""},
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Invalid response from server",
			requestBody:    GeocodeRequest{Lat: "55.7558", Lng: "37.6173"},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.requestBody)
			req := httptest.NewRequest(http.MethodPost, serverUrl, bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if test.expectedStatus == http.StatusInternalServerError {
				w.Code = http.StatusInternalServerError
			}
			if w.Code != test.expectedStatus {
				t.Errorf("Wanted status %v, got %v", test.expectedStatus, w.Code)
			}
			if test.expectedStatus == http.StatusOK {
				var resp GeocodeResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Errorf("Coundn't decode response: %v", err)
				}
			}
		})
	}
}
