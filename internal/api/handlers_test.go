package api

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charlesaraya/chirpy/internal/database"
	_ "github.com/lib/pq"
)

func loadDB(t *testing.T) (*database.Queries, error) {
	dbURL := "postgres://charlesaraya:@localhost:5432/chirpy?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Errorf("Error opening the database. got %s", err.Error())
	}
	dbQueries := database.New(db)
	return dbQueries, nil
}

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

// AssertBodyEqual checks if the response body matches a specific substring.
func assertBodyContains(t *testing.T, rec *httptest.ResponseRecorder, expected string) {
	t.Helper()
	body := rec.Body.String()
	if !strings.Contains(body, expected) {
		t.Errorf("expected body contained %q, got %q", expected, body)
	}
}

func TestHealthHandler(t *testing.T) {
	t.Run("run health handler", func(t *testing.T) {
		rec := executeRequest(t, GetHealthHandler, "GET", "/health", nil)
		assertStatus(t, rec, http.StatusOK)
		assertBodyEqual(t, rec, HealthOK)
	})
}

func TestMetricsHandler(t *testing.T) {
	cfg := &ApiConfig{}
	tmplPath := "../../" + MetricsTemplatePath
	t.Run("run server hits increment", func(t *testing.T) {
		tempDir := os.TempDir()
		os.WriteFile(filepath.Join(tempDir, "index.html"), []byte("Ok"), 0664)
		numHits := 100
		for range numHits {
			executeRequest(t, GetHomeHandler(cfg, tempDir, "/app"), "GET", "/app/", nil)
		}
		rec := executeRequest(t, GetMetricsHandler(cfg, tmplPath), "GET", "/metrics", nil)
		assertStatus(t, rec, http.StatusOK)
		assertContentType(t, rec, "text/html")
		expected := fmt.Sprintf("Chirpy has been visited %d times!", numHits)
		assertBodyContains(t, rec, expected)
	})
	t.Run("run reset metrics handler", func(t *testing.T) {
		dbQueries, _ := loadDB(t)
		resetCfg := ApiConfig{
			DBQueries: dbQueries,
			Platform:  allowedPlatform,
		}
		rec := executeRequest(t, ResetMetricsHandler(&resetCfg), "POST", "/reset", nil)
		assertStatus(t, rec, http.StatusOK)

		rec = executeRequest(t, GetMetricsHandler(&resetCfg, tmplPath), "GET", "/metrics", nil)
		expected := fmt.Sprintf("Chirpy has been visited %d times!", 0)
		assertBodyContains(t, rec, expected)
	})
}

func TestValidateChirp(t *testing.T) {
	t.Run("validate just right chirp", func(t *testing.T) {
		validChirp := strings.Repeat("chirp! ", 20)
		jsonBody := fmt.Sprintf(`{"body":"%s"}`, validChirp)
		rec := executeRequest(t, ValidateChirpHandler, "POST", "/validate_chirp", strings.NewReader(jsonBody))
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("validate too long chirp", func(t *testing.T) {
		invalidChirp := strings.Repeat("yada", 50)
		jsonBody := fmt.Sprintf(`{"body":"%s"}`, invalidChirp)
		rec := executeRequest(t, ValidateChirpHandler, "POST", "/validate_chirp", strings.NewReader(jsonBody))
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("validate invalid chirp json", func(t *testing.T) {
		jsonBody := `{"name":"Hello World!"}`
		rec := executeRequest(t, ValidateChirpHandler, "POST", "/validate_chirp", strings.NewReader(jsonBody))
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("validate empty chirp", func(t *testing.T) {
		jsonBody := `{"body":""}`
		rec := executeRequest(t, ValidateChirpHandler, "POST", "/validate_chirp", strings.NewReader(jsonBody))
		assertStatus(t, rec, http.StatusBadRequest)
	})
}
