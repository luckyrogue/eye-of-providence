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
	writer  ProfileWriter
	tokens  TokenWriter
	issuer  SessionIssuer
}

type Deps struct {
	Profile ProfileReader
	Writer  ProfileWriter
	Tokens  TokenWriter
	Issuer  SessionIssuer
}

func New(d Deps) *Service {
	return &Service{profile: d.Profile, writer: d.Writer, tokens: d.Tokens, issuer: d.Issuer}
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

func (s *Service) PatchName(ctx context.Context, userID uuid.UUID, displayName, lastName *string) error {
	if displayName == nil && lastName == nil {
		return ErrNoFields
	}
	if s.writer == nil {
		return nil
	}
	return s.writer.UpdateName(ctx, userID, displayName, lastName)
}

func (s *Service) ChangeEmail(ctx context.Context, userID uuid.UUID, email, password string, verifyFn func(hash, password string) bool) (token, newEmail string, err error) {
	if s.writer == nil {
		return "", email, nil
	}
	hash, has, err := s.writer.PasswordHash(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if !has || hash == "" {
		return "", "", ErrNoPasswordSet
	}
	if verifyFn != nil && !verifyFn(hash, password) {
		return "", "", ErrInvalidCredentials
	}
	if err := s.writer.UpdateEmail(ctx, userID, email, password); err != nil {
		return "", "", err
	}
	if s.issuer != nil {
		tok, err := s.issuer.IssueAfterCredentialChange(ctx, userID, email)
		if err != nil {
			return "", "", err
		}
		return tok, email, nil
	}
	return "", email, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, email, currentPassword, newPassword string,
	verifyFn func(hash, password string) bool, hashFn func(password string) (string, error)) (string, error) {
	if s.writer == nil {
		return "", nil
	}
	hash, has, err := s.writer.PasswordHash(ctx, userID)
	if err != nil {
		return "", err
	}
	if !has || hash == "" {
		return "", ErrNoPasswordSet
	}
	if verifyFn != nil && !verifyFn(hash, currentPassword) {
		return "", ErrInvalidCredentials
	}
	newHash, err := hashFn(newPassword)
	if err != nil {
		return "", err
	}
	if err := s.writer.UpdatePassword(ctx, userID, newHash); err != nil {
		return "", err
	}
	if s.issuer != nil {
		return s.issuer.IssueAfterCredentialChange(ctx, userID, email)
	}
	return "", nil
}
