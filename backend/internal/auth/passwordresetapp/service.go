package passwordresetapp

import (
	"context"
	"net/url"
	"strings"
	"time"
)

const defaultResetTTL = time.Hour

type Service struct {
	users    UserByEmail
	tokens   ResetTokenStore
	mail     ResetMailer
	password PasswordSetter
	tokGen   TokenGenerator
	publicURL string
	ttl      time.Duration
}

type Deps struct {
	Users     UserByEmail
	Tokens    ResetTokenStore
	Mail      ResetMailer
	Password  PasswordSetter
	TokensGen TokenGenerator
	PublicURL string
	TTL       time.Duration
}

func New(d Deps) *Service {
	ttl := d.TTL
	if ttl <= 0 {
		ttl = defaultResetTTL
	}
	return &Service{
		users: d.Users, tokens: d.Tokens, mail: d.Mail, password: d.Password,
		tokGen: d.TokensGen, publicURL: strings.TrimRight(d.PublicURL, "/"), ttl: ttl,
	}
}

func (s *Service) RequestReset(ctx context.Context, email string) error {
	if s.users == nil || s.tokens == nil || s.tokGen == nil {
		return nil
	}
	id, locale, found, err := s.users.FindByEmail(ctx, email)
	if err != nil || !found {
		return nil
	}
	tok, hash, err := s.tokGen.NewToken()
	if err != nil {
		return err
	}
	if err := s.tokens.Insert(ctx, id, hash, time.Now().Add(s.ttl)); err != nil {
		return nil
	}
	if s.mail != nil && s.publicURL != "" {
		loc := ""
		if locale != nil {
			loc = *locale
		}
		resetURL := s.publicURL + "/reset-password?token=" + url.QueryEscape(tok)
		_ = s.mail.SendReset(ctx, email, resetURL, loc)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, password string, validFn func(string) bool, hashFn func(string) (string, error)) error {
	if token == "" {
		return ErrMissingToken
	}
	if validFn != nil && !validFn(password) {
		return ErrInvalidPassword
	}
	if s.tokens == nil || s.password == nil || s.tokGen == nil {
		return nil
	}
	hash := s.tokGen.HashToken(token)
	uid, err := s.tokens.Consume(ctx, hash)
	if err != nil {
		return ErrTokenInvalid
	}
	newHash, err := hashFn(password)
	if err != nil {
		return err
	}
	return s.password.SetPassword(ctx, uid, newHash)
}
