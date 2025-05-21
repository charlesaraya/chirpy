package auth

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "testsecret"

func TestJWTs(t *testing.T) {
	validUUID := uuid.New()

	validToken, _ := MakeJWT(validUUID, testSecret, time.Hour)

	expiredToken, _ := MakeJWT(validUUID, testSecret, -time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantErr     error
		wantUUID    uuid.UUID
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			tokenSecret: testSecret,
			wantErr:     nil,
			wantUUID:    validUUID,
		},
		{
			name:        "invalid token secret",
			tokenString: validToken,
			tokenSecret: "wrong.secret",
			wantErr:     jwt.ErrSignatureInvalid,
		},
		{
			name:        "invalid token string",
			tokenString: "invalid.token.string",
			tokenSecret: testSecret,
			wantErr:     jwt.ErrTokenMalformed,
		},
		{
			name:        "empty token string",
			tokenString: "",
			tokenSecret: testSecret,
			wantErr:     jwt.ErrTokenMalformed,
		},
		{
			name:        "expired token",
			tokenString: expiredToken,
			tokenSecret: testSecret,
			wantErr:     jwt.ErrTokenExpired,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotUUID, err := ValidateJWT(tc.tokenString, tc.tokenSecret)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if err.Error() != tc.wantErr.Error() && !errors.Is(err, tc.wantErr) {
					t.Errorf("expected error '%v', got '%v'", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantUUID != gotUUID {
				t.Errorf("expected uuid %v, got %v", tc.wantUUID, gotUUID)
			}
		})
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("expected token length 64, got %d", len(token))
	}
	_, err = hex.DecodeString(token)
	if err != nil {
		t.Errorf("token is not valid hex: %v", err)
	}
}

func TestGenerateBearerToken(t *testing.T) {
	t.Run("use header with correct bearer", func(t *testing.T) {
		correctHeader := http.Header{}
		usedToken, _ := MakeRefreshToken()
		correctHeader.Add("Authorization", fmt.Sprintf("Bearer %s", usedToken))
		gotToken, _ := GetBearerToken(correctHeader)
		if usedToken != gotToken {
			t.Errorf("used token %s, got %s", usedToken, gotToken)
		}
	})
	t.Run("use header with no authorization header", func(t *testing.T) {
		emptyHeader := http.Header{}
		if _, err := GetBearerToken(emptyHeader); err.Error() != ErrMissingBearer {
			t.Errorf("expected %s error, got %s", ErrMissingBearer, err)
		}
	})
}
