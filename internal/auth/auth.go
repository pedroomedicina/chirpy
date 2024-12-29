package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	if !strings.HasPrefix(authHeader, "ApiKey ") {
		return "", errors.New("authorization header must start with 'ApiKey '")
	}

	key := strings.TrimSpace(strings.TrimPrefix(authHeader, "ApiKey "))
	if key == "" {
		return "", errors.New("api key is empty")
	}

	return key, nil
}
