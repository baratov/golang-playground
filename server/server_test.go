package server_test

import (
	"github.com/baratov/golang-playground/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HealthCheckHandler)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code is %v, but found %v", status, http.StatusOK)
	}

	expected := `{"alive":true}`
	if recorder.Body.String() != expected {
		t.Errorf("Expected body is %v, but found %v", expected, recorder.Body.String())
	}
}
