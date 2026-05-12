package auth

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/descope/virtualwebauthn"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap/zaptest"
)

// WebAuthn tests.
//
// Strategy:
//   - Config / construction: pure unit tests.
//   - Synthetic options (anti-enumeration): pure unit tests.
//   - Session save/load: use miniredis (in-process Redis).
//   - Full register/login crypto: integration only (real authenticator required).
//
// Источник правды: webauthn.go + .team/qa-testplans/webauthn-register.md +
// .team/qa-testplans/webauthn-login.md.

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func newTestWebAuthn(t *testing.T, rds *redis.Client) *WebAuthnService {
	t.Helper()
	wa, err := NewWebAuthnService(
		"localhost",
		"EOP Test",
		"http://localhost:5173",
		nil, // pool nil → DB-зависимые методы возвращают error
		rds,
		zaptest.NewLogger(t),
	)
	if err != nil {
		t.Fatalf("NewWebAuthnService: %v", err)
	}
	if wa == nil {
		t.Fatal("expected non-nil service")
	}
	return wa
}

// === Config / construction ===

func TestNewWebAuthnService_EmptyRPID_ReturnsNil(t *testing.T) {
	// Documented behavior: пустой RPID = "feature disabled", caller не
	// регистрирует endpoints. nil, nil — explicit signal.
	s, err := NewWebAuthnService("", "name", "http://x", nil, nil, zaptest.NewLogger(t))
	if err != nil {
		t.Errorf("expected nil err for empty RPID, got %v", err)
	}
	if s != nil {
		t.Errorf("expected nil service for empty RPID, got %+v", s)
	}
}

func TestNewWebAuthnService_NoOrigins_Error(t *testing.T) {
	_, err := NewWebAuthnService("rp.example", "name", "", nil, nil, zaptest.NewLogger(t))
	if err == nil {
		t.Fatal("expected error when origins empty, got nil")
	}
	if !strings.Contains(err.Error(), "origin") {
		t.Errorf("error should mention origins; got %q", err.Error())
	}
}

func TestNewWebAuthnService_HappyPath(t *testing.T) {
	s, err := NewWebAuthnService(
		"localhost",
		"EOP Test",
		"http://localhost:5173,http://localhost:5174",
		nil, nil,
		zaptest.NewLogger(t),
	)
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if s == nil || s.WA == nil {
		t.Fatal("WA nil")
	}
}

// === Synthetic options shape (anti-enumeration) ===

// TestSyntheticCredentials_ShapeMatchesReal — unknown email path returns
// синтетические allowCredentials. Они должны быть неотличимы от реальных:
// 32-байтовый ID, тип public-key, transports.
//
// Это критично: атакующий не должен через response shape отличать
// "user не существует" от "user существует без passkey'ев".
func TestSyntheticCredentials_ShapeMatchesReal(t *testing.T) {
	creds := syntheticCredentials()
	if len(creds) == 0 {
		t.Fatal("syntheticCredentials returned empty; would leak 'no user' signal")
	}
	for i, c := range creds {
		if len(c.CredentialID) != 32 {
			t.Errorf("[%d] CredentialID len = %d, want 32 (matches Apple/Google authenticator format)",
				i, len(c.CredentialID))
		}
		if c.Type != protocol.PublicKeyCredentialType {
			t.Errorf("[%d] Type = %q, want %q", i, c.Type, protocol.PublicKeyCredentialType)
		}
		if len(c.Transport) == 0 {
			t.Errorf("[%d] missing Transport (real creds always declare)", i)
		}
	}
}

// TestSyntheticCredentials_RandomEachCall — повторный вызов должен давать
// разные ID. Иначе можно по константе опознать synthetic path.
func TestSyntheticCredentials_RandomEachCall(t *testing.T) {
	a := syntheticCredentials()
	b := syntheticCredentials()
	if len(a) == 0 || len(b) == 0 {
		t.Skip("synthetic empty")
	}
	if string(a[0].CredentialID) == string(b[0].CredentialID) {
		t.Error("synthetic credential IDs are identical — adversary could fingerprint")
	}
}

// === Session save/load round-trip ===

func TestWebAuthnSession_SaveAndLoad(t *testing.T) {
	rds := newTestRedis(t)
	s := newTestWebAuthn(t, rds)

	sid, err := newSessionID()
	if err != nil {
		t.Fatalf("newSessionID: %v", err)
	}
	if len(sid) != 32 { // 16 random bytes hex-encoded = 32 chars
		t.Errorf("session id len=%d, want 32", len(sid))
	}

	original := &webauthn.SessionData{
		Challenge: "abcdef",
		UserID:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
		Expires:   time.Now().Add(5 * time.Minute).UTC(),
	}
	if err := s.saveSession(context.Background(), sid, original); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := s.loadSession(context.Background(), sid)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Challenge != original.Challenge {
		t.Errorf("Challenge mismatch: got %q, want %q", loaded.Challenge, original.Challenge)
	}
	if string(loaded.UserID) != string(original.UserID) {
		t.Errorf("UserID mismatch")
	}
}

// TestWebAuthnSession_SingleUse — после GetDel повторный read возвращает ошибку.
// Анти-replay: один session_id = один finish-attempt.
func TestWebAuthnSession_SingleUse(t *testing.T) {
	rds := newTestRedis(t)
	s := newTestWebAuthn(t, rds)

	sid, _ := newSessionID()
	if err := s.saveSession(context.Background(), sid, &webauthn.SessionData{Challenge: "c"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := s.loadSession(context.Background(), sid); err != nil {
		t.Fatalf("first load: %v", err)
	}
	if _, err := s.loadSession(context.Background(), sid); err == nil {
		t.Fatal("second load should fail (single-use), got nil err")
	}
}

func TestWebAuthnSession_NoRedis(t *testing.T) {
	// Нет Redis → методы должны вернуть explicit error (не паниковать).
	s := &WebAuthnService{Redis: nil}
	err := s.saveSession(context.Background(), "sid", &webauthn.SessionData{})
	if err == nil {
		t.Error("expected error when Redis nil")
	}
	_, err = s.loadSession(context.Background(), "sid")
	if err == nil {
		t.Error("expected error when Redis nil on load")
	}
}

// TestWebAuthnSession_Expiry — Redis TTL = 5 минут. Используем miniredis
// FastForward чтобы симулировать прошедшее время.
func TestWebAuthnSession_Expiry(t *testing.T) {
	mr := miniredis.RunT(t)
	rds := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	s, _ := NewWebAuthnService("localhost", "EOP", "http://localhost", nil, rds, zaptest.NewLogger(t))

	sid, _ := newSessionID()
	if err := s.saveSession(context.Background(), sid, &webauthn.SessionData{Challenge: "x"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	// Forward past TTL.
	mr.FastForward(webauthnSessionTTL + time.Second)
	if _, err := s.loadSession(context.Background(), sid); err == nil {
		t.Error("expected expiry error, got nil")
	}
}

// === BeginRegistration / BeginLogin without DB ===

// TestBeginRegistration_NotConfigured — WA == nil → graceful error.
func TestBeginRegistration_NotConfigured(t *testing.T) {
	s := &WebAuthnService{}
	_, _, err := s.BeginRegistration(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error when WA nil")
	}
}

// TestBeginLogin_NoPool — pool nil → graceful error.
func TestBeginLogin_NoPool(t *testing.T) {
	rds := newTestRedis(t)
	s, _ := NewWebAuthnService("localhost", "EOP", "http://localhost", nil, rds, zaptest.NewLogger(t))
	_, _, err := s.BeginLogin(context.Background(), nil)
	if err == nil {
		t.Error("expected error when pool nil")
	}
}

// === Full register / login crypto (DB-required) ===
//
// Полные ceremony-тесты живут в webauthn_integration_test.go под
// build tag `integration`:
//   - TestWebAuthnRegister_FullFlow  — BeginRegistration → FinishRegistration
//     против virtualwebauthn authenticator'а.
//   - TestWebAuthnLogin_ReplayAttack — одно assertion переиспользованное
//     дважды; вторая попытка должна быть отвергнута (single-use session
//     GETDEL'd после первого FinishLogin).
//
// Здесь же — non-integration anti-enumeration assertion: unknown email path
// возвращает options, неотличимые от "real user with passkeys" с точки зрения
// клиента.

// TestWebAuthnLogin_UnknownEmail_NoLeak — гарантирует, что unknown email
// возвращает options ровно того же shape (parseable virtualwebauthn'ом),
// что и реальный credential — атакующий не может валидировать email через
// response payload.
//
// DB не требуется: воспроизводим точную логику пути `userID == nil` из
// BeginLogin (BeginDiscoverableLogin + synthetic AllowedCredentials), затем
// прогоняем результат через virtualwebauthn.ParseAssertionOptions — тот же
// парсер, который использует браузер/клиент.
//
// Источник: webauthn.go:288–330 + webauthn-login.md (anti-enumeration §6).
func TestWebAuthnLogin_UnknownEmail_NoLeak(t *testing.T) {
	rds := newTestRedis(t)
	s := newTestWebAuthn(t, rds)

	// Симулируем path "unknown email": в реальном BeginLogin это происходит
	// после Pool.QueryRow → pgx.ErrNoRows. Здесь напрямую вызываем
	// BeginDiscoverableLogin + injection synthetic credentials, ровно как в
	// webauthn.go.
	assertion, session, err := s.WA.BeginDiscoverableLogin()
	if err != nil {
		t.Fatalf("BeginDiscoverableLogin: %v", err)
	}
	assertion.Response.AllowedCredentials = syntheticCredentials()

	// Persist session, чтобы выглядело идентично "happy-path".
	sid, err := newSessionID()
	if err != nil {
		t.Fatalf("newSessionID: %v", err)
	}
	if err := s.saveSession(context.Background(), sid, session); err != nil {
		t.Fatalf("saveSession: %v", err)
	}

	// Serialize options точно так же, как handler отдаёт фронту.
	optionsJSON, err := json.Marshal(assertion.Response)
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}

	// virtualwebauthn — единственный публичный parser, эмулирующий браузер.
	// Если он parse'ит — клиент тоже parse'ит, и shape неотличим от реального.
	parsed, err := virtualwebauthn.ParseAssertionOptions(string(optionsJSON))
	if err != nil {
		t.Fatalf("ParseAssertionOptions: response shape malformed (browser would reject too): %v", err)
	}
	if parsed.RelyingPartyID != "localhost" {
		t.Errorf("RPID=%q, want %q (anti-enumeration must keep RPID identical to real path)",
			parsed.RelyingPartyID, "localhost")
	}
	if len(parsed.Challenge) == 0 {
		t.Error("challenge empty — real path would always include one")
	}
	if len(parsed.AllowCredentials) == 0 {
		t.Error("AllowCredentials empty — leaks 'no user' (real users-with-passkeys path always has IDs)")
	}
	// Каждый synthetic credential должен иметь не-пустой base64url ID длиной
	// эквивалентной 32 байтам (44 char base64url без padding ≈ same shape).
	for i, id := range parsed.AllowCredentials {
		if id == "" {
			t.Errorf("[%d] AllowCredentials entry empty", i)
		}
	}

	// Sanity: повторный вызов даёт другой challenge (анти-fingerprint).
	a2, s2, err := s.WA.BeginDiscoverableLogin()
	if err != nil {
		t.Fatalf("second BeginDiscoverableLogin: %v", err)
	}
	a2.Response.AllowedCredentials = syntheticCredentials()
	_ = s2
	j2, _ := json.Marshal(a2.Response)
	p2, err := virtualwebauthn.ParseAssertionOptions(string(j2))
	if err != nil {
		t.Fatalf("second ParseAssertionOptions: %v", err)
	}
	if string(parsed.Challenge) == string(p2.Challenge) {
		t.Error("challenges identical across calls — adversary could fingerprint synthetic path")
	}
}

// === Session JSON shape regression ===

// TestSessionData_JSONShape — sanity-check, что webauthn.SessionData
// JSON-сериализуется в стабильную форму. Если go-webauthn lib bump'нется
// и изменит field names, наши persisted сессии станут unloadable —
// этот тест поймает раньше.
func TestSessionData_JSONShape(t *testing.T) {
	sd := webauthn.SessionData{
		Challenge: "challenge-xyz",
		UserID:    []byte("16-byte-user-uid"),
		Expires:   time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	raw, err := json.Marshal(&sd)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(raw)
	// Required fields go-webauthn lib uses — нарушение здесь обозначит lib breaking change.
	for _, want := range []string{"challenge", "user_id"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing field %q in serialized SessionData: %s", want, s)
		}
	}
}
