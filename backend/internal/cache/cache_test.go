package cache

import (
	"testing"
)

// buildOptions должен принимать оба формата:
//   - host:port (in-cluster docker)
//   - redis://user:pass@host:port (managed Redis: Dokploy, Heroku, Upstash)
// go-redis в .Addr ждёт host:port; полный URL без ParseURL приводит к
// "too many colons in address" — это и было причиной prod-инцидента.
func TestBuildOptions_HostPort(t *testing.T) {
	opts, err := buildOptions("redis:6379")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if opts.Addr != "redis:6379" {
		t.Fatalf("Addr = %q, want redis:6379", opts.Addr)
	}
	if opts.Username != "" || opts.Password != "" {
		t.Fatalf("expected no credentials, got user=%q pass=%q", opts.Username, opts.Password)
	}
}

func TestBuildOptions_RedisURLWithCredentials(t *testing.T) {
	opts, err := buildOptions("redis://default:secret@redis-host:6379/2")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if opts.Addr != "redis-host:6379" {
		t.Fatalf("Addr = %q, want redis-host:6379", opts.Addr)
	}
	if opts.Username != "default" {
		t.Fatalf("Username = %q, want default", opts.Username)
	}
	if opts.Password != "secret" {
		t.Fatalf("Password = %q, want secret", opts.Password)
	}
	if opts.DB != 2 {
		t.Fatalf("DB = %d, want 2", opts.DB)
	}
}

func TestBuildOptions_RedissTLS(t *testing.T) {
	opts, err := buildOptions("rediss://user:pwd@upstash.example:6380")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if opts.TLSConfig == nil {
		t.Fatal("rediss:// должен включать TLSConfig")
	}
}

func TestBuildOptions_TimeoutsOverridden(t *testing.T) {
	opts, err := buildOptions("redis://x:y@h:6379")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if opts.PoolSize != 10 {
		t.Fatalf("PoolSize override lost, got %d", opts.PoolSize)
	}
	if opts.DialTimeout == 0 {
		t.Fatal("DialTimeout не применён к ParseURL результату")
	}
}

func TestBuildOptions_MalformedURL(t *testing.T) {
	if _, err := buildOptions("redis://not a url"); err == nil {
		t.Fatal("expected parse error for malformed URL")
	}
}
