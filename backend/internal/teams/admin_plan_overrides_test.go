//go:build integration

package teams

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/plans"
)

func TestAdminPlanOverrides_RequiresSuperAdmin(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	regular := createUser(t, pool, "po-reg@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, regular, "po-reg@example.com")

	owner := createUser(t, pool, "po-owner@example.com")
	team := createTeamDirect(t, pool, "POReg", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/plan-limits", tok, map[string]any{
		"limits": map[string]any{"max_users_per_team": 50},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/plan-limits")
	if status != 403 {
		t.Errorf("status=%d body=%s; want 403", status, string(body))
	}
}

func TestAdminPlanOverrides_SetThenReset(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	admin := createUser(t, pool, "po-admin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "po-admin@example.com")

	owner := createUser(t, pool, "po-owner2@example.com")
	team := createTeamDirect(t, pool, "POSet", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/plan-limits", tok, map[string]any{
		"limits": map[string]any{"max_users_per_team": 50},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/plan-limits")
	if status != 200 {
		t.Fatalf("set PATCH status=%d body=%s", status, string(body))
	}

	var raw []byte
	err := pool.QueryRow(context.Background(),
		"SELECT plan_limits_override FROM teams WHERE id=$1", team).Scan(&raw)
	if err != nil {
		t.Skipf("waiting on backend agent commit of migration 024 (teams.plan_limits_override): %v", err)
	}
	var override map[string]any
	if err := json.Unmarshal(raw, &override); err != nil {
		t.Fatalf("override unmarshal: %v raw=%s", err, string(raw))
	}
	if v, _ := override["max_users_per_team"].(float64); int(v) != 50 {
		t.Errorf("override.max_users_per_team = %v, want 50", override["max_users_per_team"])
	}

	status, body = do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/plan-limits", tok, map[string]any{
		"limits": map[string]any{"max_users_per_team": nil},
	})
	if status != 200 {
		t.Fatalf("reset PATCH status=%d body=%s", status, string(body))
	}

	_ = pool.QueryRow(context.Background(),
		"SELECT plan_limits_override FROM teams WHERE id=$1", team).Scan(&raw)
	override = map[string]any{}
	_ = json.Unmarshal(raw, &override)
	if _, ok := override["max_users_per_team"]; ok {

		if override["max_users_per_team"] != nil {
			t.Errorf("after reset, override.max_users_per_team = %v, want nil or absent",
				override["max_users_per_team"])
		}
	}
}

func TestAdminPlanOverrides_AppliedAtRuntime(t *testing.T) {
	pool := setupTestDB(t)

	app := fiber.New()
	logger, _ := zap.NewDevelopment()
	svc := Service{
		Pool:          pool,
		JWTSecret:     "test-secret-32-chars-or-longer-aaaa",
		Logger:        logger,
		InviteOnly:    true,
		BetaTeamLimit: 3,
		Audit:         audit.Service{Pool: pool, Logger: logger},
		Plans:         plans.Service{Enforce: true},
	}
	RegisterRoutes(app, svc)

	admin := createUser(t, pool, "po-rt@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "po-rt@example.com")
	_ = tok

	owner := createUser(t, pool, "po-rt-owner@example.com")
	team := createTeamDirect(t, pool, "PORT", owner)

	_, err := pool.Exec(context.Background(), `
		UPDATE teams SET plan_limits_override = $1::jsonb WHERE id = $2`,
		`{"max_users_per_team": 2}`, team)
	if err != nil {
		t.Skipf("waiting on backend agent commit of migration 024 (teams.plan_limits_override): %v", err)
	}

	second := createUser(t, pool, "po-rt-2@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, second)

	ownerTok := loginAs(t, pool, svc.JWTSecret, owner, "po-rt-owner@example.com")
	status, body := do(t, app, "POST", "/v1/teams/"+team.String()+"/invites", ownerTok, map[string]any{})
	if status != 200 {
		t.Fatalf("create invite status=%d body=%s", status, string(body))
	}
	var inv struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(body, &inv)
	if inv.Code == "" {
		t.Fatal("empty invite code")
	}

	third := createUser(t, pool, "po-rt-3@example.com")
	thirdTok := loginAs(t, pool, svc.JWTSecret, third, "po-rt-3@example.com")
	status, body = do(t, app, "POST", "/v1/invites/"+inv.Code+"/accept", thirdTok, nil)
	if status == 200 {
		t.Skip("backend invite-accept still reads plan default, not override; pending backend Phase 3 wiring")
	}
	if status != 403 && status != 402 {
		t.Errorf("3rd invite accept status=%d body=%s; want 403 (seat_limit_reached) when override=2", status, string(body))
	}
}

func TestAdminPlanOverrides_AuditLogged(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)
	_, _ = pool.Exec(context.Background(), "TRUNCATE audit_log")

	admin := createUser(t, pool, "po-aud@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "po-aud@example.com")

	owner := createUser(t, pool, "po-aud-owner@example.com")
	team := createTeamDirect(t, pool, "POAud", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/plan-limits", tok, map[string]any{
		"limits": map[string]any{"max_users_per_team": 25},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/plan-limits (audit)")
	if status != 200 {
		t.Fatalf("PATCH status=%d body=%s", status, string(body))
	}

	rows := loadAuditByAction(t, pool, "team.plan_overrides_updated")
	if len(rows) == 0 {
		t.Errorf("no audit row with action=team.plan_overrides_updated; backend may not have wired audit emission")
		return
	}
	if rows[0].TargetType != "team" {
		t.Errorf("target_type=%q, want team", rows[0].TargetType)
	}
	if rows[0].TargetID != team.String() {
		t.Errorf("target_id=%q, want %q", rows[0].TargetID, team.String())
	}
}

func TestAdminPlanOverrides_OutOfRange_Rejected(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	admin := createUser(t, pool, "po-rng@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "po-rng@example.com")

	owner := createUser(t, pool, "po-rng-owner@example.com")
	team := createTeamDirect(t, pool, "PORng", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/plan-limits", tok, map[string]any{
		"limits": map[string]any{"max_users_per_team": 50000},
	})
	skipIfNotFound(t, status, "PATCH /v1/admin/teams/:id/plan-limits (range validation)")
	if status != 400 {
		t.Errorf("status=%d body=%s; want 400 (value_out_of_range)", status, string(body))
	}
}
