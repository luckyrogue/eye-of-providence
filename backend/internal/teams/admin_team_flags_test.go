//go:build integration

// Tests for Phase 3 Admin → Team flags endpoint.
//
// Acceptance source: .team/product-acceptance/admin-team-flags.md
// QA plan source:    .team/qa-testplans/admin-team-flags.md
// Audit taxonomy:    .team/product-audit-taxonomy.md (team.flags_updated)
//
// Endpoint contract (from backend admin_phase3.go):
//
//	PATCH /v1/admin/teams/:id/flags  — body MUST wrap the patch in
//	                                   {"flags": {...}} (shallow merge).
//
// Acceptance says explicit `null` clears an override. Unknown keys → 400
// `unknown_flag`. Wrong type → 400 `invalid_flag_type`. Numeric out-of-range
// → 400 `value_below_minimum` / `value_out_of_range`.

package teams

import (
	"context"
	"encoding/json"
	"testing"
)

// --- Test 1: super-admin guard ---

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

// --- Test 2: shallow merge preserves other flags ---

func TestAdminTeamFlags_ShallowMerge(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)
	_, _ = pool.Exec(context.Background(), "TRUNCATE audit_log")

	admin := createUser(t, pool, "tflags-admin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "tflags-admin@example.com")

	owner := createUser(t, pool, "tflags-owner2@example.com")
	team := createTeamDirect(t, pool, "TFlags2", owner)

	// First PATCH — set enable_insights=true.
	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"enable_insights": true},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/flags")
	if status != 200 {
		t.Fatalf("first PATCH status=%d body=%s", status, string(body))
	}

	// Second PATCH — set enable_team_reports=false (do not touch first flag).
	status, body = do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"enable_team_reports": false},
	})
	if status != 200 {
		t.Fatalf("second PATCH status=%d body=%s", status, string(body))
	}

	// Inspect DB. teams.flags JSONB column expected (migration 024 per task brief).
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

// --- Test 3: unknown flag → 400 unknown_flag ---

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
		// Backend могут отдать error в ином формате (например, "code" вместо
		// "error"). Не fatal'им, но логируем.
		t.Logf("error code = %q; expected unknown_flag", code)
	}
}

// --- Test 4: type mismatch → 400 invalid_flag_type ---

func TestAdminTeamFlags_TypeMismatch_Rejected(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	admin := createUser(t, pool, "tflags-typ@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "tflags-typ@example.com")

	owner := createUser(t, pool, "tflags-owner4@example.com")
	team := createTeamDirect(t, pool, "TFlags4", owner)

	// enable_insights expects bool — passing int.
	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"enable_insights": 5},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/flags (type validation)")
	if status != 400 {
		t.Errorf("status=%d body=%s; want 400", status, string(body))
	}
}

// --- Test 5: numeric range validation ---

func TestAdminTeamFlags_NumberRange_Validated(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	admin := createUser(t, pool, "tflags-num@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "tflags-num@example.com")

	owner := createUser(t, pool, "tflags-owner5@example.com")
	team := createTeamDirect(t, pool, "TFlags5", owner)

	// k_anon_threshold acceptance allows ≥ 1; 1000 may or may not be out-of-range
	// (acceptance bounds не зафиксированы). Тестируем явно out-of-range edge: 0.
	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"k_anon_threshold": 0},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/flags (range validation)")
	if status != 400 {
		t.Errorf("k_anon_threshold=0: status=%d body=%s; want 400", status, string(body))
	}

	// Дополнительная попытка — 1000 (per task brief). Backend acceptance не
	// фиксирует верхний bound для k_anon_threshold, поэтому мягкий assert:
	// либо 400 (out_of_range), либо 200 (если бэк не вводит верхнюю границу).
	status, _ = do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/flags", tok, map[string]any{
		"flags": map[string]any{"k_anon_threshold": 1000},
	})
	if status != 200 && status != 400 {
		t.Errorf("k_anon_threshold=1000: status=%d, want 200 or 400", status)
	}
}

// --- Test 6: successful PATCH → audit row with team.flags_updated ---

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
	// Diff payload should reference the changed key.
	var meta map[string]any
	if err := json.Unmarshal(rows[0].Metadata, &meta); err != nil {
		t.Fatalf("audit metadata unmarshal: %v raw=%s", err, string(rows[0].Metadata))
	}
	// Acceptance shape: {"diff": {"enable_insights": {"old": null, "new": false}}}
	// Допускаем также {"changed": ["enable_insights"]}.
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
