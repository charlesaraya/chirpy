// main_test.go
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func assertContentType(t *testing.T, rec *httptest.ResponseRecorder, expected string) {
	t.Helper()
	if !strings.Contains(rec.Header().Get("Content-Type"), expected) {
		t.Errorf("expected Content-Type to include %q, got %q", expected, rec.Header().Get("Content-Type"))
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
	t.Run("run server hits increment", func(t *testing.T) {
		tempDir := os.TempDir()
		os.WriteFile(filepath.Join(tempDir, "index.html"), []byte("Ok"), 0664)
		cfg := &ApiConfig{}
		numHits := 100
		for range numHits {
			executeRequest(t, GetHome(cfg, tempDir, "/app"), "GET", "/app/", nil)
		}
		rec := executeRequest(t, GetMetrics(cfg), "GET", "/metrics", nil)
		assertStatus(t, rec, http.StatusOK)
		assertContentType(t, rec, "text/plain; charset=utf-8")
		expected := fmt.Sprintf("Hits: %v", numHits)
		assertBodyEqual(t, rec, expected)
	})
	t.Run("run reset metrics handler", func(t *testing.T) {
		cfg := &ApiConfig{}
		hits := int32(42)
		cfg.ServerHits.Store(hits)

		rec := executeRequest(t, ResetMetrics(cfg), "POST", "/reset", nil)
		assertStatus(t, rec, http.StatusOK)

		rec = executeRequest(t, GetMetrics(cfg), "GET", "/metrics", nil)
		expected := fmt.Sprintf("Hits: %v", 0)
		assertBodyEqual(t, rec, expected)
	})
}
