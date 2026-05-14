package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultJWTSecret = "dev-only-secret-change-me"

type Config struct {
	Env             string
	HTTPAddr        string
	PostgresDSN     string
	ClickHouseDSN   string
	RedisAddr       string
	GeminiAPIKey    string
	GitHubClientID  string
	GitHubClientSec string
	GitHubCallback  string

	GoogleClientID  string
	GoogleClientSec string
	GoogleCallback  string

	AppleClientID   string
	AppleTeamID     string
	AppleKeyID      string
	ApplePrivateKey string
	AppleCallback   string

	WebAuthnRPID       string
	WebAuthnRPName     string
	WebAuthnOrigins    string
	JWTSecret          string
	AllowedOrigins     string
	BodyLimitBytes     int
	ReportsCronSec     int
	InviteOnly         bool
	AutoMigrate        bool
	EnableDevToken     bool
	BetaTeamLimit      int
	PlanLimitsEnforced bool
	ResendAPIKey       string
	MailFrom           string
	PublicURL          string
	VAPIDPublicKey     string
	VAPIDPrivateKey    string
	VAPIDSubject       string
}

func FromEnv() Config {
	env := getenv("EOP_ENV", "development")
	return Config{
		Env:                env,
		HTTPAddr:           resolveAddr(),
		PostgresDSN:        getenv("EOP_POSTGRES_DSN", "postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable"),
		ClickHouseDSN:      getenv("EOP_CLICKHOUSE_DSN", "clickhouse://eop:eop_dev@localhost:9000/eop"),
		RedisAddr:          getenv("EOP_REDIS_ADDR", "localhost:6379"),
		GeminiAPIKey:       os.Getenv("EOP_GEMINI_API_KEY"),
		GitHubClientID:     os.Getenv("EOP_GITHUB_CLIENT_ID"),
		GitHubClientSec:    os.Getenv("EOP_GITHUB_CLIENT_SECRET"),
		GitHubCallback:     getenv("EOP_GITHUB_CALLBACK_URL", "http://localhost:8080/v1/auth/github/callback"),
		GoogleClientID:     os.Getenv("EOP_GOOGLE_CLIENT_ID"),
		GoogleClientSec:    os.Getenv("EOP_GOOGLE_CLIENT_SECRET"),
		GoogleCallback:     getenv("EOP_GOOGLE_CALLBACK_URL", "http://localhost:8080/v1/auth/google/callback"),
		AppleClientID:      os.Getenv("EOP_APPLE_CLIENT_ID"),
		AppleTeamID:        os.Getenv("EOP_APPLE_TEAM_ID"),
		AppleKeyID:         os.Getenv("EOP_APPLE_KEY_ID"),
		ApplePrivateKey:    os.Getenv("EOP_APPLE_PRIVATE_KEY"),
		AppleCallback:      getenv("EOP_APPLE_CALLBACK_URL", "http://localhost:8080/v1/auth/apple/callback"),
		WebAuthnRPID:       os.Getenv("EOP_WEBAUTHN_RPID"),
		WebAuthnRPName:     getenv("EOP_WEBAUTHN_RPNAME", "Eye of Providence"),
		WebAuthnOrigins:    os.Getenv("EOP_WEBAUTHN_ORIGINS"),
		JWTSecret:          getenv("EOP_JWT_SECRET", defaultJWTSecret),
		AllowedOrigins:     getenv("EOP_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174,http://127.0.0.1:5173,http://127.0.0.1:5174"),
		BodyLimitBytes:     atoiOr(os.Getenv("EOP_BODY_LIMIT_BYTES"), 1<<20),
		ReportsCronSec:     atoi(os.Getenv("EOP_REPORTS_CRON_SEC")),
		InviteOnly:         getenv("EOP_INVITE_ONLY", "true") != "false",
		AutoMigrate:        boolEnv("EOP_AUTO_MIGRATE", env != "production"),
		EnableDevToken:     boolEnv("EOP_ENABLE_DEV_TOKEN", env != "production"),
		BetaTeamLimit:      atoiOr(os.Getenv("EOP_BETA_TEAM_LIMIT"), 5),
		PlanLimitsEnforced: boolEnv("EOP_PLAN_LIMITS_ENFORCED", false),
		ResendAPIKey:       os.Getenv("EOP_RESEND_API_KEY"),
		MailFrom:           getenv("EOP_MAIL_FROM", "Eye of Providence <noreply@example.com>"),
		PublicURL:          getenv("EOP_PUBLIC_URL", "http://localhost:5173"),
		VAPIDPublicKey:     os.Getenv("EOP_VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey:    os.Getenv("EOP_VAPID_PRIVATE_KEY"),
		VAPIDSubject:       getenv("EOP_VAPID_SUBJECT", "mailto:admin@example.com"),
	}
}

func (c Config) WebAuthnEnabled() bool {
	return c.WebAuthnRPID != ""
}

func (c Config) Validate() error {
	var errs []string
	if c.JWTSecret == defaultJWTSecret {
		errs = append(errs, "EOP_JWT_SECRET must be set — default secret is unsafe in any deployment")
	}
	if c.JWTSecret == "" {
		errs = append(errs, "EOP_JWT_SECRET must not be empty")
	} else if len(c.JWTSecret) < 32 {
		errs = append(errs, "EOP_JWT_SECRET must be ≥32 chars (got "+strconv.Itoa(len(c.JWTSecret))+")")
	}
	for _, o := range strings.Split(c.AllowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o == "" || o == "*" {
			continue
		}
		if strings.Contains(o, "*") {
			errs = append(errs, "EOP_ALLOWED_ORIGINS contains wildcard '*' subdomain '"+o+"' — Fiber CORS only matches exact origins")
		}
	}

	if c.WebAuthnRPID != "" && strings.TrimSpace(c.WebAuthnOrigins) == "" {
		errs = append(errs, "EOP_WEBAUTHN_ORIGINS must be set when EOP_WEBAUTHN_RPID is set")
	}

	if c.Env != "production" {
		if len(errs) == 0 {
			return nil
		}
		return errors.New("config invalid:\n  - " + strings.Join(errs, "\n  - "))
	}
	if c.AllowedOrigins == "*" {
		errs = append(errs, "EOP_ALLOWED_ORIGINS must not be '*' in production")
	}
	if c.PostgresDSN == "" {
		errs = append(errs, "EOP_POSTGRES_DSN must be set in production")
	}
	if c.ClickHouseDSN == "" {
		errs = append(errs, "EOP_CLICKHOUSE_DSN must be set in production")
	}
	if c.EnableDevToken {
		errs = append(errs, "EOP_ENABLE_DEV_TOKEN must be false in production")
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New("config invalid:\n  - " + strings.Join(errs, "\n  - "))
}

func resolveAddr() string {
	if v := os.Getenv("EOP_HTTP_ADDR"); v != "" {
		return v
	}
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":8080"
}

func atoi(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func atoiOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

func boolEnv(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "":
		return fallback
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		_, _ = fmt.Fprintf(os.Stderr, "[config] %s=%q is not a bool, using fallback=%v\n", key, v, fallback)
		return fallback
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
