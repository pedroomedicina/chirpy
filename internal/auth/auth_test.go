package auth

import "testing"

func TestHashPassword(t *testing.T) {
	password := "securepassword123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(hashedPassword) == 0 {
		t.Fatal("expected a non-empty hashed password")
	}

	t.Log("Hashed password:", hashedPassword)
}

func TestCheckPasswordHash(t *testing.T) {
	password := "securepassword123"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = CheckPasswordHash(password, hashedPassword)
	if err != nil {
		t.Fatalf("expected passwords to match, got %v", err)
	}

	err = CheckPasswordHash("wrongpassword", hashedPassword)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	t.Log("Password hashing and comparison test passed.")
}
