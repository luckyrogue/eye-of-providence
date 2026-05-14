package meapp

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

var supportedLocales = map[string]bool{"ru": true, "en": true, "kk": true, "es": true}

type Service struct {
	profile ProfileReader
	tokens  TokenWriter
}

type Deps struct {
	Profile ProfileReader
	Tokens  TokenWriter
}

func New(d Deps) *Service {
	return &Service{profile: d.Profile, tokens: d.Tokens}
}

func (s *Service) GetProfile(ctx context.Context, claims SessionClaims) (map[string]any, error) {
	out := map[string]any{
		"user_id":  claims.UserID,
		"email":    claims.Email,
		"provider": claims.Provider,
	}
	if s.profile == nil {
		return out, nil
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidSubject
	}
	ex, err := s.profile.LoadExtras(ctx, uid)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return out, nil
	}
	if ex.GithubLogin != nil {
		out["github_login"] = *ex.GithubLogin
	}
	if ex.GlobalRole != nil {
		out["global_role"] = *ex.GlobalRole
	}
	if ex.DisplayName != nil {
		out["display_name"] = *ex.DisplayName
	}
	if ex.LastName != nil {
		out["last_name"] = *ex.LastName
	}
	if ex.Phone != nil {
		out["phone"] = *ex.Phone
	}
	if ex.Locale != nil {
		out["locale"] = *ex.Locale
	}
	out["has_password"] = ex.HasPassword
	if ex.CreatedAtRFC != "" {
		out["created_at"] = ex.CreatedAtRFC
	}
	return out, nil
}

func (s *Service) PatchLocale(ctx context.Context, userID uuid.UUID, locale string) (string, error) {
	if !supportedLocales[locale] {
		return "", ErrUnsupportedLocale
	}
	if s.profile == nil {
		return locale, nil
	}
	if err := s.profile.UpdateLocale(ctx, userID, locale); err != nil {
		return "", err
	}
	return locale, nil
}

func (s *Service) ListAPITokens(ctx context.Context, userID uuid.UUID) ([]TokenRow, error) {
	if s.tokens == nil {
		return []TokenRow{}, nil
	}
	return s.tokens.List(ctx, userID)
}

type CreateAPITokenInput struct {
	Name    string
	Scope   string
	TTLDays int
}

func (s *Service) CreateAPIToken(ctx context.Context, userID uuid.UUID, in CreateAPITokenInput) (plaintext string, meta TokenRow, err error) {
	if s.tokens == nil {
		return "", TokenRow{}, ErrDBNotConfigured
	}
	name := strings.TrimSpace(in.Name)
	if len(name) > 64 {
		return "", TokenRow{}, ErrNameTooLong
	}
	scope := in.Scope
	if scope == "" {
		scope = "read"
	}
	if in.TTLDays < 0 || in.TTLDays > 365 {
		return "", TokenRow{}, ErrTTLOutOfRange
	}
	ttl := time.Duration(in.TTLDays) * 24 * time.Hour
	return s.tokens.Create(ctx, userID, name, scope, ttl)
}

func (s *Service) RevokeAPIToken(ctx context.Context, userID, tokenID uuid.UUID) (bool, error) {
	if s.tokens == nil {
		return false, ErrDBNotConfigured
	}
	return s.tokens.Revoke(ctx, userID, tokenID)
}
