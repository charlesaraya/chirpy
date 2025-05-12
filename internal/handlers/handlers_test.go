// main_test.go
package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func executeRequest(t *testing.T, handler http.HandlerFunc, method, path string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// AssertStatus checks if the response status code matches a specific code.
func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if rec.Code != expected {
		t.Errorf("expected status %d, got %d", expected, rec.Code)
	}
}

// AssertBodyEqual checks if the response body matches a specific substring.
func assertBodyEqual(t *testing.T, rec *httptest.ResponseRecorder, expected string) {
	t.Helper()
	body := rec.Body.String()
	if strings.TrimSpace(body) != expected {
		t.Errorf("expected body %q, got %q", expected, body)
	}
}

func TestHandlers(t *testing.T) {
	t.Run("run health handler", func(t *testing.T) {
		rec := executeRequest(t, GetHealth, "GET", "/health", nil)
		assertStatus(t, rec, http.StatusOK)
		assertBodyEqual(t, rec, HealthOK)
	})
}
