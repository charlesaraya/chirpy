package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTs(t *testing.T) {
	expectedTokenSecret := "ChirpyTokenSecret"
	expectedUUID := uuid.New()
	t.Run("run correct make and validate JWTs", func(t *testing.T) {
		usedTokenSecret := "ChirpyTokenSecret"
		token, _ := MakeJWT(expectedUUID, usedTokenSecret, time.Second)
		gotUUID, _ := ValidateJWT(token, expectedTokenSecret)
		if gotUUID != expectedUUID {
			t.Errorf("expected status %v, got %v", expectedUUID, gotUUID)
		}
	})
	t.Run("run invalid token secret", func(t *testing.T) {
		wrongTokenSecret := "WrongTokenSecret"
		invalidSignatureError := "signature is invalid"
		token, _ := MakeJWT(expectedUUID, expectedTokenSecret, 5*time.Second)
		_, err := ValidateJWT(token, wrongTokenSecret)
		if !strings.Contains(err.Error(), invalidSignatureError) {
			t.Errorf("expected err '%s', to contain substring '%s'", err.Error(), invalidSignatureError)
		}
	})
	t.Run("run expired token secret", func(t *testing.T) {
		expiredTokenError := "token is expired"
		token, _ := MakeJWT(expectedUUID, expectedTokenSecret, time.Second)
		time.Sleep(2 * time.Second)
		_, err := ValidateJWT(token, expectedTokenSecret)
		if !strings.Contains(err.Error(), expiredTokenError) {
			t.Errorf("expected err '%s', to contain substring '%s'", err.Error(), expiredTokenError)
		}
	})
}
