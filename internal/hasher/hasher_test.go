package hasher

import "testing"

func TestHashPassword(t *testing.T) {
	password := "password"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Errorf("Failed to hash password: %v", err)
	}
	if len(hashedPassword) == 0 {
		t.Errorf("Expected non-empty hashed password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "password"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Errorf("Failed to hash password: %v", err)
	}
	err = VerifyPassword(hashedPassword, password)
	if err != nil {
		t.Errorf("Failed to verify password: %v", err)
	}
}

func TestHashAndVerifyPassword(t *testing.T) {
	password := "password"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Errorf("Failed to hash password: %v", err)
	}
	err = VerifyPassword(hashedPassword, password)
	if err != nil {
		t.Errorf("Failed to verify password: %v", err)
	}
}

func TestHashAndVerifyPasswordWithDifferentPassword(t *testing.T) {
	password := "password"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Errorf("Failed to hash password: %v", err)
	}
	err = VerifyPassword(hashedPassword, "wrong_password")
	if err == nil {
		t.Errorf("Expected error when verifying wrong password")
	}
}
