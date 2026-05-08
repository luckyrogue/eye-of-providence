package auth

import "testing"

func TestHashAndVerify_HappyPath(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "" {
		t.Fatal("empty hash")
	}
	if !VerifyPassword(hash, "correct-horse-battery-staple") {
		t.Error("verify failed for correct password")
	}
}

func TestVerify_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("real-pass")
	if VerifyPassword(hash, "wrong-pass") {
		t.Error("verify accepted wrong password")
	}
}

func TestVerify_EmptyHash(t *testing.T) {
	if VerifyPassword("", "anything") {
		t.Error("verify accepted empty hash")
	}
}

func TestHash_DifferentEachTime(t *testing.T) {
	h1, _ := HashPassword("same-pass")
	h2, _ := HashPassword("same-pass")
	if h1 == h2 {
		t.Error("bcrypt should produce different hashes for the same password (salt)")
	}
	// Но оба должны проверяться корректно.
	if !VerifyPassword(h1, "same-pass") || !VerifyPassword(h2, "same-pass") {
		t.Error("both hashes should verify the same password")
	}
}
