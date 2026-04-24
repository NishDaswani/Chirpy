package auth

import "testing"

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "super-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == password {
		t.Fatal("HashPassword returned raw password instead of hash")
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error: %v", err)
	}
	if !match {
		t.Fatal("expected password/hash to match")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	match, err := CheckPasswordHash("wrong-password", hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error for wrong password: %v", err)
	}
	if match {
		t.Fatal("expected wrong password not to match hash")
	}
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	_, err := CheckPasswordHash("any-password", "not-a-valid-argon2id-hash")
	if err == nil {
		t.Fatal("expected error for invalid hash format")
	}
}
