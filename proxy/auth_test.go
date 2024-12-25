package main

import (
	"github.com/go-chi/jwtauth/v5"
	"testing"
)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
)

func TestRegistration(t *testing.T) {
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

	tests := []struct {
		user           User
		expectedStatus int
	}{
		{
			user:           User{Username: "Vasya123", Password: "qwerty12"},
			expectedStatus: http.StatusOK,
		},
		{
			user:           User{Username: "Vasya123", Password: ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			user:           User{Username: "", Password: "qwerty12"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			user:           User{Username: "Vasya123", Password: "qwerty12"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %v", i), func(t *testing.T) {
			body, _ := json.Marshal(test.user)
			req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Wanted status %v, got %v", test.expectedStatus, w.Code)
			}
		})
	}

}

func TestLoginSuccess(t *testing.T) {
	tokenStr := "{\"token\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdHVzZXIifQ.ds6irgKZucfY5ByDl0Vl6W87nM10BGbuCeRRLeI66eI\"}\n"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	logPass["testuser"] = string(hashedPassword)

	user := User{Username: "testuser", Password: "password123"}
	jsonData, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус-код %d, получен %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != tokenStr {
		t.Errorf("Expected token to be %v, got %v", tokenStr, w.Body.String())
	}
}

func TestLoginWrongPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	logPass["testuser"] = string(hashedPassword)
	user := User{Username: "testuser", Password: "wrongpassword"}
	jsonData, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %v, got %v", http.StatusOK, w.Code)
	}

	if w.Body.String() != "Wrong password" {
		t.Errorf("Ожидаемое сообщение об ошибке \"Wrong password\", получено: %s", w.Body.String())
	}
}

func TestLoginUserNotExist(t *testing.T) {
	user := User{Username: "nonexistentuser", Password: "password"}
	jsonData, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус-код %d, получен %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "User does not exist" {
		t.Errorf("Ожидаемое сообщение об ошибке \"User does not exist\", получено: %s", w.Body.String())
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидаемый статус-код %d, получен %d", http.StatusBadRequest, w.Code)
	}
}
