package migrate

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedFiles(t *testing.T) {
	expectPG := []string{
		"sql/postgres/001_init.up.sql",
		"sql/postgres/001_init.down.sql",
		"sql/postgres/002_teams_auth.up.sql",
		"sql/postgres/002_teams_auth.down.sql",
		"sql/postgres/003_multi_team_projects.up.sql",
		"sql/postgres/003_multi_team_projects.down.sql",
		"sql/postgres/004_subscriptions.up.sql",
		"sql/postgres/004_subscriptions.down.sql",
		"sql/postgres/005_token_version.up.sql",
		"sql/postgres/005_token_version.down.sql",

		"sql/postgres/019_user_identities.up.sql",
		"sql/postgres/019_user_identities.down.sql",
		"sql/postgres/020_webauthn_credentials.up.sql",
		"sql/postgres/020_webauthn_credentials.down.sql",
		"sql/postgres/021_team_members_observer.up.sql",
		"sql/postgres/021_team_members_observer.down.sql",

		"sql/postgres/022_email_templates.up.sql",
		"sql/postgres/022_email_templates.down.sql",
		"sql/postgres/024_team_flags_and_overrides.up.sql",
		"sql/postgres/024_team_flags_and_overrides.down.sql",

		"sql/postgres/025_content_blocks.up.sql",
		"sql/postgres/025_content_blocks.down.sql",
	}
	expectCH := []string{
		"sql/clickhouse/001_events.up.sql",
		"sql/clickhouse/001_events.down.sql",
		"sql/clickhouse/002_attribution_events.up.sql",
		"sql/clickhouse/002_attribution_events.down.sql",
	}
	for _, p := range append(expectPG, expectCH...) {
		body, err := fs.ReadFile(fsys, p)
		if err != nil {
			t.Errorf("missing embedded file %s: %v", p, err)
			continue
		}
		if len(body) == 0 {
			t.Errorf("empty embedded file %s", p)
		}
	}
}

func TestUpDownPairing(t *testing.T) {
	for _, dir := range []string{"sql/postgres", "sql/clickhouse"} {
		entries, err := fs.ReadDir(fsys, dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		ups := map[string]bool{}
		downs := map[string]bool{}
		for _, e := range entries {
			name := e.Name()
			switch {
			case strings.HasSuffix(name, ".up.sql"):
				ups[strings.TrimSuffix(name, ".up.sql")] = true
			case strings.HasSuffix(name, ".down.sql"):
				downs[strings.TrimSuffix(name, ".down.sql")] = true
			}
		}
		for k := range ups {
			if !downs[k] {
				t.Errorf("%s/%s.up.sql has no matching .down.sql", dir, k)
			}
		}
		for k := range downs {
			if !ups[k] {
				t.Errorf("%s/%s.down.sql has no matching .up.sql", dir, k)
			}
		}
	}
}

func TestRunPostgresEmptyDSN(t *testing.T) {
	if err := RunPostgres(t.Context(), ""); err != nil {
		t.Errorf("expected nil for empty dsn, got: %v", err)
	}
	if err := RunClickHouse(t.Context(), ""); err != nil {
		t.Errorf("expected nil for empty dsn, got: %v", err)
	}
}

func TestNewPostgresEmptyDSN(t *testing.T) {
	if _, err := NewPostgres(""); err == nil {
		t.Error("expected error for empty dsn")
	}
	if _, err := NewClickHouse(""); err == nil {
		t.Error("expected error for empty dsn")
	}
}
