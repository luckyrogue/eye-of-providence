package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	webauthnSessionTTL    = 5 * time.Minute
	webauthnSessionPrefix = "webauthn:session:"
)

type WebAuthnService struct {
	WA     *webauthn.WebAuthn
	Pool   *pgxpool.Pool
	Redis  *redis.Client
	Logger *zap.Logger
}

func NewWebAuthnService(rpid, rpName, origins string, pool *pgxpool.Pool, rds *redis.Client, log *zap.Logger) (*WebAuthnService, error) {
	if rpid == "" {
		return nil, nil
	}
	rps := []string{}
	for _, o := range strings.Split(origins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			rps = append(rps, o)
		}
	}
	if len(rps) == 0 {
		return nil, errors.New("webauthn: at least one origin required")
	}
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          rpid,
		RPDisplayName: rpName,
		RPOrigins:     rps,
	})
	if err != nil {
		return nil, fmt.Errorf("webauthn config: %w", err)
	}
	return &WebAuthnService{WA: wa, Pool: pool, Redis: rds, Logger: log}, nil
}

type webauthnUser struct {
	id          uuid.UUID
	name        string
	displayName string
	creds       []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte          { return u.id[:] }
func (u *webauthnUser) WebAuthnName() string        { return u.name }
func (u *webauthnUser) WebAuthnDisplayName() string { return u.displayName }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.creds
}

func (s *WebAuthnService) loadUser(ctx context.Context, userID uuid.UUID) (*webauthnUser, error) {
	u := &webauthnUser{id: userID}
	if s.Pool == nil {
		return u, nil
	}

	var email, name *string
	_ = s.Pool.QueryRow(ctx,
		`SELECT email, display_name FROM users WHERE id = $1`, userID,
	).Scan(&email, &name)
	if email != nil {
		u.name = *email
	}
	if name != nil {
		u.displayName = *name
	}
	if u.displayName == "" {
		u.displayName = u.name
	}

	rows, err := s.Pool.Query(ctx, `
		SELECT credential_id, public_key, aaguid, sign_count, transports
		FROM webauthn_credentials
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			credID, pubKey []byte
			aaguid         *uuid.UUID
			signCount      uint32
			transports     string
		)
		if err := rows.Scan(&credID, &pubKey, &aaguid, &signCount, &transports); err != nil {
			return nil, err
		}
		c := webauthn.Credential{
			ID:        credID,
			PublicKey: pubKey,
			Authenticator: webauthn.Authenticator{
				SignCount: signCount,
			},
		}
		if aaguid != nil {
			c.Authenticator.AAGUID = aaguid[:]
		}
		if transports != "" {
			for _, t := range strings.Split(transports, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					c.Transport = append(c.Transport, protocol.AuthenticatorTransport(t))
				}
			}
		}
		u.creds = append(u.creds, c)
	}
	return u, rows.Err()
}

func newSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *WebAuthnService) saveSession(ctx context.Context, sid string, data *webauthn.SessionData) error {
	if s.Redis == nil {
		return errors.New("webauthn: redis required")
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.Redis.Set(ctx, webauthnSessionPrefix+sid, raw, webauthnSessionTTL).Err()
}

func (s *WebAuthnService) loadSession(ctx context.Context, sid string) (*webauthn.SessionData, error) {
	if s.Redis == nil {
		return nil, errors.New("webauthn: redis required")
	}
	raw, err := s.Redis.GetDel(ctx, webauthnSessionPrefix+sid).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("webauthn: session not found or expired")
	}
	if err != nil {
		return nil, err
	}
	var data webauthn.SessionData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *WebAuthnService) BeginRegistration(ctx context.Context, userID uuid.UUID) (*protocol.CredentialCreation, string, error) {
	if s.WA == nil {
		return nil, "", errors.New("webauthn: not configured")
	}
	u, err := s.loadUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	creation, session, err := s.WA.BeginRegistration(u)
	if err != nil {
		return nil, "", err
	}
	sid, err := newSessionID()
	if err != nil {
		return nil, "", err
	}
	if err := s.saveSession(ctx, sid, session); err != nil {
		return nil, "", err
	}
	return creation, sid, nil
}

func (s *WebAuthnService) FinishRegistration(ctx context.Context, userID uuid.UUID, sid string, body []byte, nickname string) error {
	if s.WA == nil {
		return errors.New("webauthn: not configured")
	}
	if s.Pool == nil {
		return errors.New("webauthn: db required")
	}
	session, err := s.loadSession(ctx, sid)
	if err != nil {
		return err
	}
	u, err := s.loadUser(ctx, userID)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	cred, err := s.WA.FinishRegistration(u, *session, req)
	if err != nil {
		return err
	}

	var transports []string
	for _, t := range cred.Transport {
		transports = append(transports, string(t))
	}
	var aaguid *uuid.UUID
	if len(cred.Authenticator.AAGUID) == 16 {
		g, err := uuid.FromBytes(cred.Authenticator.AAGUID)
		if err == nil {
			aaguid = &g
		}
	}
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO webauthn_credentials
			(user_id, credential_id, public_key, aaguid, sign_count, transports, nickname)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))
	`, userID, cred.ID, cred.PublicKey, aaguid, int64(cred.Authenticator.SignCount), strings.Join(transports, ","), strings.TrimSpace(nickname))
	return err
}

func (s *WebAuthnService) BeginLogin(ctx context.Context, email *string) (*protocol.CredentialAssertion, string, error) {
	if s.WA == nil {
		return nil, "", errors.New("webauthn: not configured")
	}
	if s.Pool == nil {
		return nil, "", errors.New("webauthn: db required")
	}

	var userID *uuid.UUID
	if email != nil {
		var uid uuid.UUID
		err := s.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, *email).Scan(&uid)
		if err == nil {
			userID = &uid
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, "", err
		}
	}

	if userID == nil {

		assertion, session, err := s.WA.BeginDiscoverableLogin()
		if err != nil {
			return nil, "", err
		}

		assertion.Response.AllowedCredentials = syntheticCredentials()
		sid, err := newSessionID()
		if err != nil {
			return nil, "", err
		}

		if err := s.saveSession(ctx, sid, session); err != nil {
			return nil, "", err
		}
		return assertion, sid, nil
	}

	u, err := s.loadUser(ctx, *userID)
	if err != nil {
		return nil, "", err
	}
	if len(u.creds) == 0 {

		assertion, session, err := s.WA.BeginDiscoverableLogin()
		if err != nil {
			return nil, "", err
		}
		assertion.Response.AllowedCredentials = syntheticCredentials()
		sid, err := newSessionID()
		if err != nil {
			return nil, "", err
		}
		if err := s.saveSession(ctx, sid, session); err != nil {
			return nil, "", err
		}
		return assertion, sid, nil
	}

	assertion, session, err := s.WA.BeginLogin(u)
	if err != nil {
		return nil, "", err
	}
	sid, err := newSessionID()
	if err != nil {
		return nil, "", err
	}
	if err := s.saveSession(ctx, sid, session); err != nil {
		return nil, "", err
	}
	return assertion, sid, nil
}

func (s *WebAuthnService) FinishLogin(ctx context.Context, sid string, body []byte) (uuid.UUID, error) {
	if s.WA == nil {
		return uuid.Nil, errors.New("webauthn: not configured")
	}
	if s.Pool == nil {
		return uuid.Nil, errors.New("webauthn: db required")
	}
	session, err := s.loadSession(ctx, sid)
	if err != nil {
		return uuid.Nil, err
	}
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	if err != nil {
		return uuid.Nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if len(session.UserID) == 0 {
		var resolvedUserID uuid.UUID
		handler := func(rawID, userHandle []byte) (webauthn.User, error) {

			if len(userHandle) != 16 {
				return nil, errors.New("invalid user handle")
			}
			uid, err := uuid.FromBytes(userHandle)
			if err != nil {
				return nil, err
			}
			u, err := s.loadUser(ctx, uid)
			if err != nil {
				return nil, err
			}
			if len(u.creds) == 0 {
				return nil, errors.New("user has no credentials")
			}
			resolvedUserID = uid
			return u, nil
		}
		cred, err := s.WA.FinishDiscoverableLogin(handler, *session, req)
		if err != nil {
			return uuid.Nil, err
		}
		if err := s.updateSignCount(ctx, cred); err != nil {
			s.Logger.Warn("webauthn sign_count update failed", zap.Error(err))
		}
		return resolvedUserID, nil
	}

	if len(session.UserID) != 16 {
		return uuid.Nil, errors.New("invalid session user id")
	}
	uid, err := uuid.FromBytes(session.UserID)
	if err != nil {
		return uuid.Nil, err
	}
	u, err := s.loadUser(ctx, uid)
	if err != nil {
		return uuid.Nil, err
	}
	cred, err := s.WA.FinishLogin(u, *session, req)
	if err != nil {
		return uuid.Nil, err
	}
	if err := s.updateSignCount(ctx, cred); err != nil {
		s.Logger.Warn("webauthn sign_count update failed", zap.Error(err))
	}
	return uid, nil
}

func (s *WebAuthnService) updateSignCount(ctx context.Context, cred *webauthn.Credential) error {
	if s.Pool == nil {
		return nil
	}
	_, err := s.Pool.Exec(ctx, `
		UPDATE webauthn_credentials
		SET sign_count = $1, last_used_at = now()
		WHERE credential_id = $2
	`, int64(cred.Authenticator.SignCount), cred.ID)
	return err
}

func syntheticCredentials() []protocol.CredentialDescriptor {

	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return []protocol.CredentialDescriptor{
		{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: b,
			Transport:    []protocol.AuthenticatorTransport{protocol.Internal, protocol.Hybrid},
		},
	}
}

type PasskeyRow struct {
	ID         uuid.UUID  `json:"id"`
	Nickname   string     `json:"nickname,omitempty"`
	AAGUID     string     `json:"aaguid,omitempty"`
	Transports []string   `json:"transports,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

func (s *WebAuthnService) ListPasskeys(ctx context.Context, userID uuid.UUID) ([]PasskeyRow, error) {
	if s.Pool == nil {
		return nil, nil
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT id, COALESCE(nickname, ''), aaguid, COALESCE(transports, ''), created_at, last_used_at
		FROM webauthn_credentials
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PasskeyRow
	for rows.Next() {
		var (
			pk     PasskeyRow
			aaguid *uuid.UUID
			tr     string
			nick   string
		)
		if err := rows.Scan(&pk.ID, &nick, &aaguid, &tr, &pk.CreatedAt, &pk.LastUsedAt); err != nil {
			return nil, err
		}
		pk.Nickname = nick
		if aaguid != nil {
			pk.AAGUID = aaguid.String()
		}
		if tr != "" {
			for _, t := range strings.Split(tr, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					pk.Transports = append(pk.Transports, t)
				}
			}
		}
		out = append(out, pk)
	}
	return out, rows.Err()
}

func (s *WebAuthnService) PasskeyCredentialIDForUser(ctx context.Context, userID, passkeyID uuid.UUID) ([]byte, error) {
	if s.Pool == nil {
		return nil, errors.New("webauthn: db required")
	}
	var cid []byte
	err := s.Pool.QueryRow(ctx,
		`SELECT credential_id FROM webauthn_credentials WHERE id = $1 AND user_id = $2`,
		passkeyID, userID,
	).Scan(&cid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPasskeyNotFound
	}
	return cid, err
}

func (s *WebAuthnService) DeletePasskey(ctx context.Context, userID, passkeyID uuid.UUID) error {
	if s.Pool == nil {
		return errors.New("webauthn: db required")
	}
	res, err := s.Pool.Exec(ctx,
		`DELETE FROM webauthn_credentials WHERE id = $1 AND user_id = $2`,
		passkeyID, userID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrPasskeyNotFound
	}
	return nil
}

var ErrPasskeyNotFound = errors.New("passkey not found")

func EncodeChallenge(c protocol.URLEncodedBase64) string {
	return base64.RawURLEncoding.EncodeToString(c)
}
