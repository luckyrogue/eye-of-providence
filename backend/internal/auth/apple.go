package auth

import (
	"context"
	"errors"
	"net/url"
)

// AppleOAuth — Apple Sign-in (OIDC-compatible) provider. Phase 1 SKELETON ONLY.
//
// Apple отличается от GitHub/Google тем, что client_secret — не статическая
// строка, а ECDSA P-256 JWT, который мы должны подписывать сами (TTL ≤ 6 мес):
//
//	header  = {alg: ES256, kid: <AppleKeyID>}
//	payload = {iss: <AppleTeamID>, iat, exp, aud: "https://appleid.apple.com", sub: <AppleClientID>}
//
// Также Apple отдаёт name/email **только при первом login'е** — нужен Redis
// cache на 5 min, чтобы Settings-страница смогла подхватить имя.
//
// Phase 2 implements:
//   - generateClientSecret(): JWT-sign через ECDSA P-256 (golang.org/x/crypto или jwt/v5)
//   - JWKS fetch + cache: https://appleid.apple.com/auth/keys
//   - id_token verification (alg=RS256 от Apple, kid из header)
//   - email_verified guard (Apple возвращает 'email_verified' как string "true")
//   - Redis cache name/email на первой регистрации
//
// Phase 1: stub реализация — методы возвращают "not implemented", чтобы
// инициализация в main.go (когда мы её добавим) проходила компиляцию.
type AppleOAuth struct {
	TeamID      string // Apple Developer Team ID (из Apple Developer console)
	KeyID       string // Apple "Sign in with Apple" Key ID
	ClientID    string // Apple Services ID (например, "com.eop.web")
	PrivateKey  string // PEM-encoded ECDSA P-256 private key (создаётся в Apple Dev console)
	CallbackURL string // redirect_uri
}

// статический check — *AppleOAuth удовлетворяет OAuthProvider.
var _ OAuthProvider = (*AppleOAuth)(nil)

func NewAppleOAuth(teamID, keyID, clientID, privateKey, callbackURL string) *AppleOAuth {
	return &AppleOAuth{
		TeamID:      teamID,
		KeyID:       keyID,
		ClientID:    clientID,
		PrivateKey:  privateKey,
		CallbackURL: callbackURL,
	}
}

// Name — provider key для interface OAuthProvider.
func (a *AppleOAuth) Name() string { return "apple" }

// AuthCodeURL — Apple OAuth authorize URL. Apple-specific параметры:
//   - response_mode=form_post (Apple отправляет callback как POST, не GET)
//   - scope=name email (запрашивается явно; Apple отдаёт name только первый раз)
//
// Phase 1: формирует URL корректно — этого достаточно для frontend redirect.
// Реальный flow завершит Phase 2 (Exchange).
func (a *AppleOAuth) AuthCodeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", a.ClientID)
	q.Set("redirect_uri", a.CallbackURL)
	q.Set("response_type", "code")
	q.Set("scope", "name email")
	q.Set("response_mode", "form_post")
	q.Set("state", state)
	return "https://appleid.apple.com/auth/authorize?" + q.Encode()
}

// Exchange — Phase 1 STUB. TODO Phase 2:
//  1. generateClientSecret(): ECDSA P-256 JWT signed with PrivateKey/KeyID/TeamID,
//     iss=TeamID, sub=ClientID, aud="https://appleid.apple.com", exp ≤ 6 months
//  2. POST https://appleid.apple.com/auth/token с form-encoded:
//     {client_id, client_secret, code, grant_type=authorization_code, redirect_uri}
//  3. Получить id_token (RS256), verify через JWKS https://appleid.apple.com/auth/keys
//  4. Извлечь sub, email, email_verified (Apple возвращает email_verified как string)
//  5. Если name пришёл в исходном callback form-post — закешировать в Redis
//     по subject на 5 минут (Apple возвращает name только при первой регистрации)
func (a *AppleOAuth) Exchange(_ context.Context, _ string) (*ExternalUser, error) {
	// Stub: provider зарегистрирован, но flow не реализован. Возвращаем явную
	// ошибку чтобы handler.go вернул 502 Bad Gateway / 503 Unavailable, а не
	// панику или JWT на несуществующего юзера.
	return nil, errors.New("apple sign-in: phase 1 stub, exchange not implemented yet")
}
