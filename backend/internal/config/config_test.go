package config

import (
	"strings"
	"testing"
)

func TestValidate_DevelopmentAlwaysOK(t *testing.T) {
	c := Config{Env: "development", JWTSecret: defaultJWTSecret, AllowedOrigins: "*"}
	if err := c.Validate(); err != nil {
		t.Fatalf("dev should not fail validation, got %v", err)
	}
}

func TestValidate_ProductionRejectsDefaults(t *testing.T) {
	c := Config{
		Env:            "production",
		JWTSecret:      defaultJWTSecret,
		AllowedOrigins: "*",
		PostgresDSN:    "",
		ClickHouseDSN:  "",
	}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected validation error in production with default JWT, got nil")
	}
	for _, want := range []string{"EOP_JWT_SECRET", "EOP_ALLOWED_ORIGINS", "EOP_POSTGRES_DSN", "EOP_CLICKHOUSE_DSN"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("validation error missing %q: %v", want, err)
		}
	}
}

func TestValidate_ProductionAcceptsRealValues(t *testing.T) {
	c := Config{
		Env:            "production",
		JWTSecret:      strings.Repeat("a", 64),
		AllowedOrigins: "https://example.com",
		PostgresDSN:    "postgres://x@y:5432/z",
		ClickHouseDSN:  "clickhouse://x@y:9000/z",
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected no error with real values, got %v", err)
	}
}

func TestValidate_ProductionShortSecretRejected(t *testing.T) {
	c := Config{
		Env:            "production",
		JWTSecret:      "short",
		AllowedOrigins: "https://example.com",
		PostgresDSN:    "postgres://x@y:5432/z",
		ClickHouseDSN:  "clickhouse://x@y:9000/z",
	}
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "≥32") {
		t.Fatalf("expected ≥32 chars error, got %v", err)
	}
}
