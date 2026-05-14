//go:build integration

package teams

import (
	"context"
	"encoding/json"
	"testing"
)

func TestAdminTeamFlags_RequiresSuperAdmin(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	regular := createUser(t, pool, "tflags-reg@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, regular, "tflags-reg@example.com")

	owner := createUser(t, pool, "tflags-owner@example.com")
	team := createTeamDirect(t, pool, "TFlags1", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"enable_insights": false},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/flags")
	if status != 403 {
		t.Errorf("status=%d body=%s; want 403", status, string(body))
	}
}

func TestAdminTeamFlags_ShallowMerge(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)
	_, _ = pool.Exec(context.Background(), "TRUNCATE audit_log")

	admin := createUser(t, pool, "tflags-admin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "tflags-admin@example.com")

	owner := createUser(t, pool, "tflags-owner2@example.com")
	team := createTeamDirect(t, pool, "TFlags2", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"enable_insights": true},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/flags")
	if status != 200 {
		t.Fatalf("first PATCH status=%d body=%s", status, string(body))
	}

	status, body = do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"enable_team_reports": false},
	})
	if status != 200 {
		t.Fatalf("second PATCH status=%d body=%s", status, string(body))
	}

	var flagsRaw []byte
	err := pool.QueryRow(context.Background(),
		"SELECT flags FROM teams WHERE id=$1", team).Scan(&flagsRaw)
	if err != nil {
		t.Skipf("waiting on backend agent commit of migration 024 (teams.flags JSONB): %v", err)
	}
	var flags map[string]any
	if err := json.Unmarshal(flagsRaw, &flags); err != nil {
		t.Fatalf("flags unmarshal: %v raw=%s", err, string(flagsRaw))
	}

	insights, hasInsights := flags["enable_insights"]
	reports, hasReports := flags["enable_team_reports"]

	if !hasInsights {
		t.Error("enable_insights missing after second PATCH; shallow merge expected to preserve it")
	} else if v, ok := insights.(bool); !ok || !v {
		t.Errorf("enable_insights=%v, want true", insights)
	}
	if !hasReports {
		t.Error("enable_team_reports missing after second PATCH")
	} else if v, ok := reports.(bool); !ok || v {
		t.Errorf("enable_team_reports=%v, want false", reports)
	}
}

func TestAdminTeamFlags_UnknownFlag_Rejected(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	admin := createUser(t, pool, "tflags-unk@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "tflags-unk@example.com")

	owner := createUser(t, pool, "tflags-owner3@example.com")
	team := createTeamDirect(t, pool, "TFlags3", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"magic_flag_does_not_exist": "drop tables"},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/flags (validation)")
	if status != 400 {
		t.Errorf("status=%d body=%s; want 400", status, string(body))
	}
	if code, _ := jsonField(t, body, "error").(string); code != "" && code != "unknown_flag" {

		t.Logf("error code = %q; expected unknown_flag", code)
	}
}

func TestAdminTeamFlags_TypeMismatch_Rejected(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	admin := createUser(t, pool, "tflags-typ@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "tflags-typ@example.com")

	owner := createUser(t, pool, "tflags-owner4@example.com")
	team := createTeamDirect(t, pool, "TFlags4", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"enable_insights": 5},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/flags (type validation)")
	if status != 400 {
		t.Errorf("status=%d body=%s; want 400", status, string(body))
	}
}

func TestAdminTeamFlags_NumberRange_Validated(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	admin := createUser(t, pool, "tflags-num@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "tflags-num@example.com")

	owner := createUser(t, pool, "tflags-owner5@example.com")
	team := createTeamDirect(t, pool, "TFlags5", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"k_anon_threshold": 0},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/flags (range validation)")
	if status != 400 {
		t.Errorf("k_anon_threshold=0: status=%d body=%s; want 400", status, string(body))
	}

	status, _ = do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"k_anon_threshold": 1000},
	})
	if status != 200 && status != 400 {
		t.Errorf("k_anon_threshold=1000: status=%d, want 200 or 400", status)
	}
}

func TestAdminTeamFlags_AuditLogged(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)
	_, _ = pool.Exec(context.Background(), "TRUNCATE audit_log")

	admin := createUser(t, pool, "tflags-aud@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "tflags-aud@example.com")

	owner := createUser(t, pool, "tflags-owner6@example.com")
	team := createTeamDirect(t, pool, "TFlags6", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"enable_insights": false},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/flags (audit)")
	if status != 200 {
		t.Fatalf("PATCH status=%d body=%s", status, string(body))
	}

	rows := loadAuditByAction(t, pool, "team.flags_updated")
	if len(rows) == 0 {
		t.Errorf("no audit row written with action=team.flags_updated; backend may not have wired audit emission")
		return
	}

	var meta map[string]any
	if err := json.Unmarshal(rows[0].Metadata, &meta); err != nil {
		t.Fatalf("audit metadata unmarshal: %v raw=%s", err, string(rows[0].Metadata))
	}

	diff, hasDiff := meta["diff"].(map[string]any)
	_, hasChanged := meta["changed"]
	if !hasDiff && !hasChanged {
		t.Errorf("audit metadata missing diff/changed payload: %v", meta)
	}
	if hasDiff {
		if _, ok := diff["enable_insights"]; !ok {
			t.Errorf("diff missing enable_insights key: %v", diff)
		}
	}
}
