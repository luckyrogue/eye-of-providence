package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/httperr"
)

const (
	tokenTTL = 14 * 24 * time.Hour

	// handoffCookieName — HttpOnly cookie со встроенным short-lived JWT (TTL
	// 30s). После OAuth/passkey callback'а backend ставит cookie + redirect
	// на frontend `/auth/complete`. Frontend читает cookie через
	// /v1/me/session-handoff (one-shot), получает JWT, сохраняет в localStorage.
	// Cookie + handoff endpoint — потому что URL fragment `#token=...` светит
	// JWT в browser history + referrer headers.
	handoffCookieName = "eop_session_handoff"
	handoffTTL        = 30 * time.Second
)

// Service — DI bundle для auth-handler'ов. GitHub оставлен ради legacy
// callback URL (см. RegisterRoutes). Providers — generic map для unified flow.
type Service struct {
	JWTSecret string
	// Providers — карта включённых OAuth-провайдеров. Конструируется в main.go
	// на основе env: если EOP_GOOGLE_CLIENT_ID пустой — провайдер не попадает
	// в map → endpoints для google не регистрируются.
	Providers      map[string]OAuthProvider
	GitHub         *GitHubOAuth // deprecated: остался для обратной совместимости с тестами и явных вызовов
	Logger         *zap.Logger
	Users          *UsersPG
	Pool           *pgxpool.Pool
	WebAuthn       *WebAuthnService // nil если WebAuthn выключен (см. config.WebAuthnEnabled)
	PublicURL      string           // base URL фронта для redirect'ов (например "http://localhost:5173")
	SecureCookies  bool             // true → cookie с Secure флагом (prod)
	EnableDevToken bool
}

// ProvidersList — упорядоченный slice provider name'ов для /v1/auth/config.
func (s Service) ProvidersList() []string {
	out := make([]string, 0, len(s.Providers))
	for name := range s.Providers {
		out = append(out, name)
	}
	// Стабильный порядок: github, google, apple (если включён).
	order := []string{"github", "google", "apple"}
	sorted := make([]string, 0, len(out))
	for _, k := range order {
		for _, n := range out {
			if n == k {
				sorted = append(sorted, n)
				break
			}
		}
	}
	// Любые остальные (passkey/sso не в этом списке) — append после.
	for _, n := range out {
		found := false
		for _, k := range order {
			if n == k {
				found = true
				break
			}
		}
		if !found {
			sorted = append(sorted, n)
		}
	}
	return sorted
}

func RegisterRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/auth")

	if s.EnableDevToken {
		registerDevToken(g, s)
	}

	// OAuth providers — generic flow. Каждый зарегистрирует:
	//   GET /v1/auth/<name>/login    — sets state cookie, redirects
	//   GET /v1/auth/<name>/callback — exchange code, set handoff cookie, redirect
	for name, p := range s.Providers {
		registerOAuthProvider(g, s, name, p)
	}

	// Backward compatibility: легаси код в main.go всё ещё может передать
	// s.GitHub, но без записи в s.Providers (старый путь). Регистрируем
	// дополнительный handler ТОЛЬКО если "github" не в Providers — иначе
	// будет double-register и Fiber запаникует.
	if _, has := s.Providers["github"]; !has && s.GitHub != nil {
		registerOAuthProvider(g, s, "github", s.GitHub)
	}

	// Session handoff — public endpoint, читает HttpOnly cookie, отдаёт JWT,
	// очищает cookie (one-shot).
	g.Get("/session-handoff-internal", handleSessionHandoff(s)) // backwards path; primary lives at /v1/me/session-handoff

	// WebAuthn / Passkey endpoints (registration требует auth, login public).
	if s.WebAuthn != nil {
		// Public — login.
		g.Post("/webauthn/login/begin", handleWebAuthnLoginBegin(s))
		g.Post("/webauthn/login/finish", handleWebAuthnLoginFinish(s))
		// Authed — registration. Используем Middleware напрямую (нельзя через
		// app.Group("/v1/auth", Middleware...) — это сломает другие public
		// endpoints в этой группе).
		mw := Middleware(s.JWTSecret, s.Pool)
		g.Post("/webauthn/register/begin", mw, handleWebAuthnRegisterBegin(s))
		g.Post("/webauthn/register/finish", mw, handleWebAuthnRegisterFinish(s))
	}
}

// RegisterSessionHandoffRoute — /v1/me/session-handoff registered ВНЕ /v1/me
// middleware-группы (читает cookie напрямую без Bearer). Вызывается из main.go
// ДО auth.RegisterMeRoutes чтобы handler не попал под Middleware.
func RegisterSessionHandoffRoute(app *fiber.App, s Service) {
	app.Get("/v1/me/session-handoff", handleSessionHandoff(s))
}

// registerOAuthProvider — generic OAuth dance. Login → state cookie + redirect.
// Callback → state-mismatch guard → exchange → upsert identity → handoff cookie
// + 302 redirect к фронту.
func registerOAuthProvider(g fiber.Router, s Service, name string, p OAuthProvider) {
	g.Get("/"+name+"/login", func(c *fiber.Ctx) error {
		state := randomState()
		// return_to (опционально) — куда фронт вернётся после complete.
		// Сохраняем в state cookie рядом со random nonce: "<nonce>|<return_to>"
		// (если задан). Cookie sigend в смысле HMAC не нужен, потому что
		// secret = nonce, мы сверяем equality.
		returnTo := c.Query("return_to")
		stateVal := state
		if returnTo != "" {
			stateVal = state + "|" + returnTo
		}
		c.Cookie(&fiber.Cookie{
			Name:     "eop_oauth_state",
			Value:    stateVal,
			HTTPOnly: true,
			SameSite: "Lax",
			Secure:   s.SecureCookies,
			MaxAge:   600,
		})
		return c.Redirect(p.AuthCodeURL(state))
	})

	g.Get("/"+name+"/callback", func(c *fiber.Ctx) error {
		got := c.Query("state")
		stored := c.Cookies("eop_oauth_state")
		// Извлекаем return_to (если был).
		nonce := stored
		returnTo := ""
		if i := strings.Index(stored, "|"); i > 0 {
			nonce = stored[:i]
			returnTo = stored[i+1:]
		}
		if got == "" || got != nonce {
			return httperr.BadRequest(c, "state_mismatch", "state mismatch")
		}
		code := c.Query("code")
		if code == "" {
			// Apple/OAuth могут вернуть error=access_denied — выдадим стабильный код.
			if c.Query("error") == "access_denied" {
				return httperr.BadRequest(c, "user_denied", "user denied access")
			}
			return httperr.BadRequest(c, "missing_code", "missing code")
		}
		user, err := p.Exchange(c.Context(), code)
		if err != nil {
			s.Logger.Warn("oauth exchange failed", zap.String("provider", name), zap.Error(err))
			return httperr.BadGateway(c, "oauth_exchange_failed", "oauth exchange failed")
		}
		if user.Email == "" {
			return httperr.BadRequest(c, "email_not_verified", "provider did not return verified email")
		}

		userUUID, linkErr := upsertOAuthUser(c.Context(), s, name, user)
		if linkErr != nil {
			s.Logger.Warn("oauth link failed", zap.String("provider", name), zap.Error(linkErr))
			if errors.Is(linkErr, errIdentityLinkRequired) {
				// Email collision policy: existing password-user без email_verified=true
				// → 409 "sign in to link". Согласно decisions-confirmed §4, мы линкуем
				// автоматически если provider gave email_verified=true (что обязательно
				// для Google/GitHub по нашим guard'ам), иначе требуем re-auth.
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "identity_link_required",
					"detail": "an account exists with this email; sign in with password to link",
					"email": user.Email,
					"provider": name,
				})
			}
			return httperr.Conflict(c, "oauth_link_failed", "could not link account")
		}

		// Выпускаем JWT и кладём в handoff cookie. Frontend получит через
		// /v1/me/session-handoff (one-shot).
		tv, _ := TokenVersion(c.Context(), s.Pool, userUUID)
		tok, err := IssueJWT(s.JWTSecret, userUUID.String(), user.Email, name, tv, tokenTTL)
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return httperr.Internal(c)
		}
		setHandoffCookie(c, s, tok)
		// Clear oauth state cookie.
		c.Cookie(&fiber.Cookie{Name: "eop_oauth_state", Value: "", MaxAge: -1, HTTPOnly: true, Secure: s.SecureCookies})

		// Redirect к /auth/complete. PublicURL хранит origin фронта без trailing slash.
		dest := strings.TrimRight(s.PublicURL, "/") + "/auth/complete"
		if returnTo != "" {
			dest += "?return_to=" + base64URLEncode(returnTo)
		}
		return c.Redirect(dest, fiber.StatusFound)
	})
}

// upsertOAuthUser — реализует email-collision policy (decisions-confirmed §4):
//  1. Сначала ищем по (provider, subject) в user_identities → существующий
//     mapping → login (вернём user_id).
//  2. Если нет: ищем users.email == email. Найден → link identity к существующему
//     юзеру (auto-link, потому что provider гарантировал email_verified=true в Exchange).
//  3. Не найден: create user + identity (новый пользователь).
//
// Возвращает (userID, error). errIdentityLinkRequired — sentinel для cases когда
// мы хотим вернуть 409 (Phase 1 не использует, оставлен на будущее для unverified-email path).
func upsertOAuthUser(ctx context.Context, s Service, provider string, ext *ExternalUser) (uuid.UUID, error) {
	if s.Pool == nil {
		// In-memory fallback — derive deterministic UUID, no persistence.
		return uuid.NewSHA1(uuid.NameSpaceURL, []byte(provider+":"+ext.Subject)), nil
	}

	// Step 1: existing identity?
	var existingUserID uuid.UUID
	err := s.Pool.QueryRow(ctx, `
		SELECT user_id FROM user_identities WHERE provider = $1 AND subject = $2
	`, provider, ext.Subject).Scan(&existingUserID)
	if err == nil {
		// Existing identity — sync email (verified провайдером) on the user
		// row только если он сейчас пустой.
		_, _ = s.Pool.Exec(ctx, `
			UPDATE users SET email = COALESCE(NULLIF(users.email, ''), $1) WHERE id = $2
		`, ext.Email, existingUserID)
		return existingUserID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}

	// Step 2: existing user by email?
	var linkedUserID uuid.UUID
	err = s.Pool.QueryRow(ctx,
		`SELECT id FROM users WHERE email = $1`, ext.Email,
	).Scan(&linkedUserID)
	if err == nil {
		// Link: write identity row to existing user. Provider guarantee'ит
		// email_verified=true (см. github.go/google.go Exchange security guards),
		// поэтому auto-link безопасен.
		_, err := s.Pool.Exec(ctx, `
			INSERT INTO user_identities (user_id, provider, subject, email)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (provider, subject) DO NOTHING
		`, linkedUserID, provider, ext.Subject, ext.Email)
		if err != nil {
			return uuid.Nil, err
		}
		return linkedUserID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}

	// Step 3: create new user + identity (атомарно через tx).
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	newID := uuid.New()
	displayName := ext.Name
	if displayName == "" {
		displayName = ext.Login
	}
	if displayName == "" {
		displayName = ext.Email
	}
	githubLogin := ""
	if provider == "github" {
		githubLogin = ext.Login
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, github_login, display_name)
		VALUES ($1, $2, NULLIF($3, ''), $4)
		ON CONFLICT (id) DO NOTHING
	`, newID, ext.Email, githubLogin, displayName); err != nil {
		return uuid.Nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_identities (user_id, provider, subject, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, subject) DO NOTHING
	`, newID, provider, ext.Subject, ext.Email); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return newID, nil
}

// errIdentityLinkRequired — sentinel для будущего, когда добавим
// "unverified email" path. В Phase 2 все провайдеры гарантируют email_verified,
// поэтому return value не используется.
var errIdentityLinkRequired = errors.New("identity link requires re-auth")

// setHandoffCookie — кладёт one-shot JWT в HttpOnly cookie. Lax SameSite —
// потому что callback это redirect от другого origin'а (provider) к нашему
// домену, Strict бы дропнул.
func setHandoffCookie(c *fiber.Ctx, s Service, tok string) {
	c.Cookie(&fiber.Cookie{
		Name:     handoffCookieName,
		Value:    tok,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   s.SecureCookies,
		MaxAge:   int(handoffTTL.Seconds()),
	})
}

// handleSessionHandoff — GET /v1/me/session-handoff.
// Читает cookie, верифицирует JWT (signature + exp), возвращает {token}, clear'ит cookie.
//
// Один-к-одному: каждый callback кладёт фрешный JWT, frontend читает РОВНО ОДИН
// раз. Повторное чтение получит пустой cookie → 404.
func handleSessionHandoff(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tok := c.Cookies(handoffCookieName)
		if tok == "" {
			return httperr.NotFound(c, "no_handoff", "no pending session handoff")
		}
		// Clear cookie regardless of validity (one-shot semantics).
		// fasthttp only serializes Max-Age when MaxAge>0, so use Expires in the past
		// (RFC 6265) — matches integration test and browser deletion semantics.
		c.Cookie(&fiber.Cookie{
			Name:     handoffCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   0,
			Expires:  time.Unix(0, 0).UTC(),
			HTTPOnly: true,
			SameSite: "Lax",
			Secure:   s.SecureCookies,
		})
		claims, err := ParseJWT(s.JWTSecret, tok)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_handoff", "invalid handoff token")
		}
		// Дополнительный sanity check: токен только что выпущен (≤handoffTTL).
		// Если iat старше — это попытка реплея устаревшего cookie.
		if claims.IssuedAt != nil {
			if time.Since(claims.IssuedAt.Time) > handoffTTL+5*time.Second {
				return httperr.Unauthorized(c, "expired_handoff", "handoff token too old")
			}
		}
		return c.JSON(fiber.Map{"token": tok})
	}
}

// RegisterIdentitiesRoutes — auth'd endpoints для /v1/me/identities (list / delete).
// Регистрируется отдельно от RegisterMeRoutes (тот в auth.MeService с другим scope).
func RegisterIdentitiesRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/me", Middleware(s.JWTSecret, s.Pool))
	g.Get("/identities", handleListIdentities(s))
	g.Delete("/identities/:id", handleDeleteIdentity(s))

	if s.WebAuthn != nil {
		g.Get("/passkeys", handleListPasskeys(s))
		g.Delete("/passkeys/:id", handleDeletePasskey(s))
	}
}

// IdentityRow — DTO для GET /v1/me/identities.
type IdentityRow struct {
	ID         uuid.UUID  `json:"id"`
	Provider   string     `json:"provider"`
	Subject    string     `json:"subject"`
	Email      string     `json:"email,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

func handleListIdentities(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		if s.Pool == nil {
			return c.JSON(fiber.Map{"identities": []IdentityRow{}})
		}
		rows, err := s.Pool.Query(c.Context(), `
			SELECT id, provider, subject, COALESCE(email, ''), created_at
			FROM user_identities
			WHERE user_id = $1
			ORDER BY created_at DESC
		`, uid)
		if err != nil {
			return httperr.Internal(c)
		}
		defer rows.Close()
		out := []IdentityRow{}
		for rows.Next() {
			var r IdentityRow
			if err := rows.Scan(&r.ID, &r.Provider, &r.Subject, &r.Email, &r.CreatedAt); err != nil {
				return httperr.Internal(c)
			}
			out = append(out, r)
		}
		return c.JSON(fiber.Map{"identities": out})
	}
}

func handleDeleteIdentity(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		idStr := c.Params("id")
		identityID, err := uuid.Parse(idStr)
		if err != nil {
			return httperr.BadRequest(c, "invalid_id", "invalid identity id")
		}
		if s.Pool == nil {
			return httperr.Unavailable(c, "db_required", "identities require database")
		}

		// Owner-check одновременно с lockout-guard.
		factors, err := CountAuthFactors(c.Context(), s.Pool, uid, &identityID, nil)
		if err != nil {
			return s.internalErr(c, err)
		}
		if factors.Total() == 0 {
			// После удаления юзер не сможет войти ни одним способом.
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "last_auth_factor",
				"detail":  "cannot remove last sign-in method; set a password or add a passkey first",
			})
		}

		res, err := s.Pool.Exec(c.Context(),
			`DELETE FROM user_identities WHERE id = $1 AND user_id = $2`,
			identityID, uid,
		)
		if err != nil {
			return s.internalErr(c, err)
		}
		if res.RowsAffected() == 0 {
			return httperr.NotFound(c, "not_found", "identity not found")
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func handleListPasskeys(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		rows, err := s.WebAuthn.ListPasskeys(c.Context(), uid)
		if err != nil {
			return s.internalErr(c, err)
		}
		return c.JSON(fiber.Map{"passkeys": rows})
	}
}

func handleDeletePasskey(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		idStr := c.Params("id")
		passkeyID, err := uuid.Parse(idStr)
		if err != nil {
			return httperr.BadRequest(c, "invalid_id", "invalid passkey id")
		}

		// Получаем credential_id raw для exclude в CountAuthFactors.
		credID, err := s.WebAuthn.PasskeyCredentialIDForUser(c.Context(), uid, passkeyID)
		if err != nil {
			if errors.Is(err, ErrPasskeyNotFound) {
				return httperr.NotFound(c, "not_found", "passkey not found")
			}
			return s.internalErr(c, err)
		}

		factors, err := CountAuthFactors(c.Context(), s.Pool, uid, nil, credID)
		if err != nil {
			return s.internalErr(c, err)
		}
		if factors.Total() == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "last_auth_factor",
				"detail": "cannot remove last sign-in method; set a password or add an identity first",
			})
		}

		if err := s.WebAuthn.DeletePasskey(c.Context(), uid, passkeyID); err != nil {
			if errors.Is(err, ErrPasskeyNotFound) {
				return httperr.NotFound(c, "not_found", "passkey not found")
			}
			return s.internalErr(c, err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// ---------------------------------------------------------------------------
// WebAuthn handlers

func handleWebAuthnRegisterBegin(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		creation, sid, err := s.WebAuthn.BeginRegistration(c.Context(), uid)
		if err != nil {
			s.Logger.Warn("webauthn begin register failed", zap.Error(err))
			return httperr.BadRequest(c, "webauthn_begin_failed", err.Error())
		}
		return c.JSON(fiber.Map{
			"options":    creation,
			"session_id": sid,
		})
	}
}

type webauthnFinishRegisterReq struct {
	SessionID   string          `json:"session_id"`
	Attestation json.RawMessage `json:"attestation"`
	Nickname    string          `json:"nickname"`
}

func handleWebAuthnRegisterFinish(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		var req webauthnFinishRegisterReq
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		if req.SessionID == "" || len(req.Attestation) == 0 {
			return httperr.BadRequest(c, "missing_fields", "session_id and attestation required")
		}
		if err := s.WebAuthn.FinishRegistration(c.Context(), uid, req.SessionID, req.Attestation, req.Nickname); err != nil {
			s.Logger.Warn("webauthn finish register failed", zap.Error(err))
			return httperr.BadRequest(c, "webauthn_finish_failed", err.Error())
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

type webauthnLoginBeginReq struct {
	Email *string `json:"email"`
}

func handleWebAuthnLoginBegin(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req webauthnLoginBeginReq
		// Empty body допустим — это discoverable / usernameless login.
		_ = c.BodyParser(&req)
		assertion, sid, err := s.WebAuthn.BeginLogin(c.Context(), req.Email)
		if err != nil {
			s.Logger.Warn("webauthn begin login failed", zap.Error(err))
			return httperr.BadRequest(c, "webauthn_begin_failed", err.Error())
		}
		return c.JSON(fiber.Map{
			"options":    assertion,
			"session_id": sid,
		})
	}
}

type webauthnLoginFinishReq struct {
	SessionID string          `json:"session_id"`
	Assertion json.RawMessage `json:"assertion"`
	// ReturnTo — куда фронт хочет вернуться после /auth/complete. Опционально.
	ReturnTo string `json:"return_to"`
}

func handleWebAuthnLoginFinish(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req webauthnLoginFinishReq
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		if req.SessionID == "" || len(req.Assertion) == 0 {
			return httperr.BadRequest(c, "missing_fields", "session_id and assertion required")
		}
		uid, err := s.WebAuthn.FinishLogin(c.Context(), req.SessionID, req.Assertion)
		if err != nil {
			s.Logger.Warn("webauthn finish login failed", zap.Error(err))
			return httperr.Unauthorized(c, "webauthn_finish_failed", err.Error())
		}
		// Достаём email для JWT.
		var email string
		if s.Pool != nil {
			_ = s.Pool.QueryRow(c.Context(),
				`SELECT COALESCE(email, '') FROM users WHERE id = $1`, uid,
			).Scan(&email)
		}
		tv, _ := TokenVersion(c.Context(), s.Pool, uid)
		tok, err := IssueJWT(s.JWTSecret, uid.String(), email, "passkey", tv, tokenTTL)
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return httperr.Internal(c)
		}
		// Возвращаем JSON-токен напрямую (passkey-login не идёт через external
		// redirect — это XHR на нашем домене, нет необходимости в handoff cookie
		// dance). Но для единообразия фронта можем дополнительно поставить cookie.
		setHandoffCookie(c, s, tok)
		return c.JSON(fiber.Map{"token": tok, "user_id": uid.String()})
	}
}

// internalErr — единая точка для 500-ответов из этого пакета (handler.go).
func (s Service) internalErr(c *fiber.Ctx, err error) error {
	rid, _ := c.Locals("requestid").(string)
	s.Logger.Error("internal error",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()),
		zap.String("rid", rid),
		zap.Error(err),
	)
	return httperr.Internal(c)
}

// registerDevToken — выдаёт JWT без OAuth (для локальной разработки).
// Регистрируется только когда EnableDevToken=true (по умолчанию выключен в production).
func registerDevToken(g fiber.Router, s Service) {
	g.Post("/dev-token", func(c *fiber.Ctx) error {
		userIDStr := c.Query("user_id")
		var userID uuid.UUID
		if userIDStr == "" {
			userID = uuid.New()
		} else {
			parsed, err := uuid.Parse(userIDStr)
			if err != nil {
				return httperr.BadRequest(c, "invalid_user_id", "user_id must be uuid")
			}
			userID = parsed
		}
		email := fmt.Sprintf("dev-%s@local.eop", userID.String()[:8])
		if err := s.Users.Upsert(c.Context(), userID, email, ""); err != nil {
			s.Logger.Warn("user upsert failed (continuing with token)", zap.Error(err))
		}
		tv, _ := TokenVersion(c.Context(), s.Pool, userID)
		tok, err := IssueJWT(s.JWTSecret, userID.String(), email, "dev", tv, tokenTTL)
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"token": tok, "user_id": userID.String()})
	})
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func base64URLEncode(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
