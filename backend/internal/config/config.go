package config

import (
	"os"
	"strconv"
)

type Config struct {
	Env             string
	HTTPAddr        string
	PostgresDSN     string
	ClickHouseDSN   string
	RedisAddr       string
	GeminiAPIKey    string
	GitHubClientID  string
	GitHubClientSec string
	JWTSecret       string
	ReportsCronSec  int // 0 = выключено
}

func FromEnv() Config {
	return Config{
		Env:             getenv("EOP_ENV", "development"),
		HTTPAddr:        getenv("EOP_HTTP_ADDR", ":8080"),
		PostgresDSN:     getenv("EOP_POSTGRES_DSN", "postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable"),
		ClickHouseDSN:   getenv("EOP_CLICKHOUSE_DSN", "clickhouse://eop:eop_dev@localhost:9000/eop"),
		RedisAddr:       getenv("EOP_REDIS_ADDR", "localhost:6379"),
		GeminiAPIKey:    os.Getenv("EOP_GEMINI_API_KEY"),
		GitHubClientID:  os.Getenv("EOP_GITHUB_CLIENT_ID"),
		GitHubClientSec: os.Getenv("EOP_GITHUB_CLIENT_SECRET"),
		JWTSecret:       getenv("EOP_JWT_SECRET", "dev-only-secret-change-me"),
		ReportsCronSec:  atoi(os.Getenv("EOP_REPORTS_CRON_SEC")),
	}
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

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
