package auth

import (
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := time.Minute * 15

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(token) == 0 {
		t.Fatal("Expected a non-empty token")
	}

	t.Logf("Generated JWT: %s", token)
}

func TestMakeJWTInvalidSecret(t *testing.T) {
	userID := uuid.New()
	tokenSecret := ""
	expiresIn := time.Minute * 15

	_, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err == nil {
		t.Fatal("Expected an error for empty secret, got none")
	}
}

func TestValidateJWTValid(t *testing.T) {
	// Create a valid token
	tokenSecret := "test-secret"
	userID := uuid.New()
	token, _ := MakeJWT(userID, tokenSecret, time.Minute*15)

	// Validate the token
	parsedID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if parsedID != userID {
		t.Fatalf("Expected user ID %s, got %s", userID, parsedID)
	}

	t.Log("Token validation succeeded")
}

func TestValidateJWTExpired(t *testing.T) {
	tokenSecret := "test-secret"
	userID := uuid.New()
	token, _ := MakeJWT(userID, tokenSecret, -time.Minute) // Token already expired

	_, err := ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Fatal("Expected an error for expired token, got none")
	}

	t.Logf("Expected error: %v", err)
}

func TestGetBearerTokenValid(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer valid_token_string")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token != "valid_token_string" {
		t.Fatalf("Expected token 'valid_token_string', got '%s'", token)
	}
}

func TestGetBearerTokenMissingHeader(t *testing.T) {
	headers := http.Header{}

	_, err := GetBearerToken(headers)
	if err == nil || err.Error() != "authorization header is missing" {
		t.Fatalf("Expected error 'authorization header is missing', got %v", err)
	}
}

func TestGetBearerTokenInvalidPrefix(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Invalid valid_token_string")

	_, err := GetBearerToken(headers)
	if err == nil || err.Error() != "authorization header must start with 'Bearer '" {
		t.Fatalf("Expected error 'authorization header must start with 'Bearer '', got %v", err)
	}
}

func TestGetBearerTokenEmptyToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer ")

	_, err := GetBearerToken(headers)
	if err == nil || err.Error() != "authorization token is empty" {
		t.Fatalf("Expected error 'authorization token is empty', got %v", err)
	}
}
