package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestSignPayload_Deterministic(t *testing.T) {
	body := []byte(`{"event":"commit.ingested"}`)
	a := signPayload("k", body)
	b := signPayload("k", body)
	if a != b {
		t.Errorf("non-deterministic: %s vs %s", a, b)
	}
}

func TestSignPayload_DifferentKey_DifferentSig(t *testing.T) {
	body := []byte(`{"event":"x"}`)
	if signPayload("k1", body) == signPayload("k2", body) {
		t.Error("same sig with different keys")
	}
}

func TestVerifySignature_HappyPath(t *testing.T) {
	secret := "whk_abc"
	body := []byte("hello")
	sig := signPayload(secret, body)
	if !VerifySignature(secret, body, "sha256="+sig) {
		t.Error("verify failed for valid sig")
	}
}

func TestVerifySignature_BadKey(t *testing.T) {
	body := []byte("hello")
	sig := signPayload("right", body)
	if VerifySignature("wrong", body, "sha256="+sig) {
		t.Error("verify accepted wrong key")
	}
}

func TestVerifySignature_BadFormat(t *testing.T) {
	cases := []string{
		"",
		"sha256=notHex",
		"md5=" + signPayload("k", []byte("x")),
		signPayload("k", []byte("x")), // без префикса
	}
	for _, h := range cases {
		if VerifySignature("k", []byte("x"), h) {
			t.Errorf("verify accepted bad header: %q", h)
		}
	}
}

func TestVerifySignature_BodyTampering(t *testing.T) {
	secret := "k"
	original := []byte(`{"a":1}`)
	tampered := []byte(`{"a":2}`)
	sig := signPayload(secret, original)
	if VerifySignature(secret, tampered, "sha256="+sig) {
		t.Error("verify accepted tampered body")
	}
}

func TestSignPayload_Format(t *testing.T) {
	sig := signPayload("k", []byte("x"))
	if len(sig) != 64 {
		t.Errorf("sig len=%d, want 64 (sha256 hex)", len(sig))
	}
	if _, err := hex.DecodeString(sig); err != nil {
		t.Errorf("sig not hex: %v", err)
	}
}

func TestSignPayload_KnownVector(t *testing.T) {
	// Known answer: HMAC-SHA256("key", "data") = 5031fe3d989c6d1537a013fa6e739da23463fdaec3b70137d828e36ace221bd0
	mac := hmac.New(sha256.New, []byte("key"))
	mac.Write([]byte("data"))
	want := hex.EncodeToString(mac.Sum(nil))
	got := signPayload("key", []byte("data"))
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestGenerateSecret_Unique(t *testing.T) {
	a, _ := generateSecret()
	b, _ := generateSecret()
	if a == b {
		t.Error("generateSecret produced collision")
	}
	if !strings.HasPrefix(a, "whk_") {
		t.Errorf("secret %q missing whk_ prefix", a)
	}
	if len(a) != 4+64 { // whk_ + 32 bytes hex
		t.Errorf("secret len=%d, want 68", len(a))
	}
}

func TestValidEvents(t *testing.T) {
	if !validEvents[EventCommitIngested] {
		t.Error("commit.ingested should be valid")
	}
	if !validEvents[EventReportGenerated] {
		t.Error("report.generated should be valid")
	}
	if validEvents["random.event"] {
		t.Error("random.event should be invalid")
	}
}
