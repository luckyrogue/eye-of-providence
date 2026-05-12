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
	GitHubCallback  string // полный URL https://api.../v1/auth/github/callback
	JWTSecret       string
	AllowedOrigins  string // CORS, comma-separated, "*" разрешено
	BodyLimitBytes  int    // глобальный лимит тела запроса
	ReportsCronSec  int    // 0 = выключено
	InviteOnly      bool   // true = регистрация только по invite (первый user всегда может)
	AutoMigrate     bool   // true = прогнать .sql миграции на старте
	EnableDevToken  bool   // true = роут /v1/auth/dev-token зарегистрирован
	BetaTeamLimit   int    // максимальное число команд в бете (0 = без лимита). Super_admin может больше.
	// PlanLimitsEnforced — kill-switch для feature gates (plans package).
	// false (default в бете) = все проверки no-op'ятся, всем доступны все фичи.
	// true (на GA) = enforces Limits.MaxUsersPerTeam, MaxWebhooks, SSO, AuditLog.
	PlanLimitsEnforced bool
	ResendAPIKey    string // если пустой — Mailer = Noop (только лог, без HTTP-вызовов)
	MailFrom        string // RFC-5322 from-address: `Eye of Providence <noreply@app.example.com>`
	PublicURL       string // base URL дашборда для ссылок в email'ах
	// Web Push (PWA notifications) — VAPID ECDSA P-256 keypair. Генерируется
	// один раз через `cmd/vapid-gen`. Если public/private пусты — push
	// endpoints возвращают 503 "push not configured".
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string // RFC 8292: "mailto:admin@..." or app URL
}

func FromEnv() Config {
	env := getenv("EOP_ENV", "development")
	return Config{
		Env:             env,
		HTTPAddr:        resolveAddr(),
		PostgresDSN:     getenv("EOP_POSTGRES_DSN", "postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable"),
		ClickHouseDSN:   getenv("EOP_CLICKHOUSE_DSN", "clickhouse://eop:eop_dev@localhost:9000/eop"),
		RedisAddr:       getenv("EOP_REDIS_ADDR", "localhost:6379"),
		GeminiAPIKey:    os.Getenv("EOP_GEMINI_API_KEY"),
		GitHubClientID:  os.Getenv("EOP_GITHUB_CLIENT_ID"),
		GitHubClientSec: os.Getenv("EOP_GITHUB_CLIENT_SECRET"),
		GitHubCallback:  getenv("EOP_GITHUB_CALLBACK_URL", "http://localhost:8080/v1/auth/github/callback"),
		JWTSecret:       getenv("EOP_JWT_SECRET", defaultJWTSecret),
		AllowedOrigins:  getenv("EOP_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174,http://127.0.0.1:5173,http://127.0.0.1:5174"),
		BodyLimitBytes:  atoiOr(os.Getenv("EOP_BODY_LIMIT_BYTES"), 1<<20), // 1 MiB
		ReportsCronSec:  atoi(os.Getenv("EOP_REPORTS_CRON_SEC")),
		InviteOnly:      getenv("EOP_INVITE_ONLY", "true") != "false",
		AutoMigrate:     boolEnv("EOP_AUTO_MIGRATE", env != "production"),
		EnableDevToken:  boolEnv("EOP_ENABLE_DEV_TOKEN", env != "production"),
		BetaTeamLimit:   atoiOr(os.Getenv("EOP_BETA_TEAM_LIMIT"), 3),
		PlanLimitsEnforced: boolEnv("EOP_PLAN_LIMITS_ENFORCED", false),
		ResendAPIKey:    os.Getenv("EOP_RESEND_API_KEY"),
		MailFrom:        getenv("EOP_MAIL_FROM", "Eye of Providence <noreply@example.com>"),
		PublicURL:       getenv("EOP_PUBLIC_URL", "http://localhost:5173"),
		VAPIDPublicKey:  os.Getenv("EOP_VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: os.Getenv("EOP_VAPID_PRIVATE_KEY"),
		VAPIDSubject:    getenv("EOP_VAPID_SUBJECT", "mailto:admin@example.com"),
	}
}

// Validate — проверки, которые должны провалить запуск.
//
// Hardening: проверки JWT (default-secret, длина ≥32) теперь применяются
// УНИВЕРСАЛЬНО, не только в production. Раньше dev-build с дефолтным
// секретом мог по ошибке улететь на staging/prod через мисдеплой и
// принимать любой подделанный JWT. Защитимся явно — fail fast.
func (c Config) Validate() error {
	var errs []string
	// Эти проверки работают во всех env, чтобы случайно поднятый "dev" билд
	// в публичной сети не оказался открытым.
	if c.JWTSecret == defaultJWTSecret {
		errs = append(errs, "EOP_JWT_SECRET must be set — default secret is unsafe in any deployment")
	}
	if c.JWTSecret == "" {
		errs = append(errs, "EOP_JWT_SECRET must not be empty")
	} else if len(c.JWTSecret) < 32 {
		errs = append(errs, "EOP_JWT_SECRET must be ≥32 chars (got "+strconv.Itoa(len(c.JWTSecret))+")")
	}
	// CORS: запрещаем wildcard subdomain (Fiber CORS поддерживает только exact match,
	// "https://*.foo.com" молча НЕ матчится — даём явную ошибку при старте).
	for _, o := range strings.Split(c.AllowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o == "" || o == "*" {
			continue
		}
		if strings.Contains(o, "*") {
			errs = append(errs, "EOP_ALLOWED_ORIGINS contains wildcard '*' subdomain '"+o+"' — Fiber CORS only matches exact origins")
		}
	}

	if c.Env != "production" {
		if len(errs) == 0 {
			return nil
		}
		return errors.New("config invalid:\n  - " + strings.Join(errs, "\n  - "))
	}
	// Production-only проверки (JWT checks выше уже сделаны).
	if c.AllowedOrigins == "*" {
		errs = append(errs, "EOP_ALLOWED_ORIGINS must not be '*' in production")
	}
	if c.PostgresDSN == "" {
		errs = append(errs, "EOP_POSTGRES_DSN must be set in production")
	}
	if c.ClickHouseDSN == "" {
		errs = append(errs, "EOP_CLICKHOUSE_DSN must be set in production")
	}
	// dev-token endpoint в production = catastrophic — выдаёт JWT любому.
	if c.EnableDevToken {
		errs = append(errs, "EOP_ENABLE_DEV_TOKEN must be false in production")
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New("config invalid:\n  - " + strings.Join(errs, "\n  - "))
}

// resolveAddr — приоритет:
//  1. EOP_HTTP_ADDR (наш контракт, ":8080" формат)
//  2. PORT (Render/Heroku/Railway/Fly auto-inject port number)
//  3. ":8080" default
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
		// Не паникуем — но печатаем понятное сообщение в stderr,
		// чтобы опечатка в env не игнорировалась молча.
		fmt.Fprintf(os.Stderr, "[config] %s=%q is not a bool, using fallback=%v\n", key, v, fallback)
		return fallback
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
