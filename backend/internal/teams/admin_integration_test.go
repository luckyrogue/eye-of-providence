//go:build integration

package teams

import (
	"context"
	"encoding/json"
	"testing"
)

func TestAdmin_NonSuper_403(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	uid := createUser(t, pool, "regular@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, uid, "regular@example.com")

	for _, path := range []string{
		"/v1/admin/stats",
		"/v1/admin/teams",
		"/v1/admin/users",
	} {
		status, _ := do(t, app, "GET", path, tok, nil)
		if status != 403 {
			t.Errorf("path %q: status = %d, want 403", path, status)
		}
	}
}

func TestAdminStats_HappyPath(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "admin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin@example.com")

	owner := createUser(t, pool, "owner@example.com")
	t1 := createTeamDirect(t, pool, "T1", owner)
	_ = createTeamDirect(t, pool, "T2", owner)

	member := createUser(t, pool, "member@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", t1, member)

	status, body := do(t, app, "GET", "/v1/admin/stats", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}
	var out struct {
		UsersTotal   int `json:"users_total"`
		TeamsTotal   int `json:"teams_total"`
		MembersTotal int `json:"members_total"`
		BetaLimit    int `json:"beta_limit"`
	}
	_ = json.Unmarshal(body, &out)
	if out.UsersTotal < 3 {
		t.Errorf("users_total = %d, expected ≥3", out.UsersTotal)
	}
	if out.TeamsTotal != 2 {
		t.Errorf("teams_total = %d, want 2", out.TeamsTotal)
	}
	if out.BetaLimit != 3 {
		t.Errorf("beta_limit = %d, want 3", out.BetaLimit)
	}
}

func TestAdminListTeams(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "admin2@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin2@example.com")

	owner := createUser(t, pool, "owner2@example.com")
	createTeamDirect(t, pool, "Acme", owner)

	status, body := do(t, app, "GET", "/v1/admin/teams", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var out struct {
		Teams []map[string]any `json:"teams"`
	}
	_ = json.Unmarshal(body, &out)
	if len(out.Teams) == 0 {
		t.Fatal("teams list пустой")
	}
	found := false
	for _, team := range out.Teams {
		if team["name"] == "Acme" {
			found = true
			if team["owner_email"] != "owner2@example.com" {
				t.Errorf("owner_email = %v", team["owner_email"])
			}
		}
	}
	if !found {
		t.Error("Acme team не найдена в выдаче")
	}
}

func TestAdminListUsers(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "admin3@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin3@example.com")

	createUser(t, pool, "user1@example.com")
	createUser(t, pool, "user2@example.com")

	status, body := do(t, app, "GET", "/v1/admin/users", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var out struct {
		Users []map[string]any `json:"users"`
	}
	_ = json.Unmarshal(body, &out)
	if len(out.Users) < 3 {
		t.Errorf("users count = %d, want ≥3", len(out.Users))
	}
}

func TestAdminDeleteUser(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "admin4@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin4@example.com")

	victim := createUser(t, pool, "victim@example.com")

	status, _ := do(t, app, "DELETE", "/v1/admin/users/"+victim.String(), tok, nil)
	if status != 200 {
		t.Fatalf("delete status=%d", status)
	}

	var count int
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM users WHERE id=$1", victim).Scan(&count)
	if count != 0 {
		t.Error("user не удалён")
	}
}

func TestAdminDeleteUser_CannotDeleteSelf(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "selfdel@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "selfdel@example.com")

	status, _ := do(t, app, "DELETE", "/v1/admin/users/"+admin.String(), tok, nil)
	if status == 200 {
		t.Error("admin удалил себя — должно быть 403/409")
	}
}

func TestAdminUpdateUser_PromoteToSuper(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "admin5@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin5@example.com")

	target := createUser(t, pool, "tobeAdmin@example.com")
	status, _ := do(t, app, "PATCH", "/v1/admin/users/"+target.String(), tok, map[string]string{
		"global_role": "super_admin",
	})
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var role string
	_ = pool.QueryRow(context.Background(),
		"SELECT global_role FROM users WHERE id=$1", target).Scan(&role)
	if role != "super_admin" {
		t.Errorf("global_role = %q, want super_admin", role)
	}
}

func TestBetaInfo_PublicEndpoint(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	uid := createUser(t, pool, "any@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, uid, "any@example.com")

	owner := createUser(t, pool, "ownerB@example.com")
	createTeamDirect(t, pool, "B1", owner)

	status, body := do(t, app, "GET", "/v1/beta/info", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var info struct {
		Limit          int `json:"limit"`
		TeamsCount     int `json:"teams_count"`
		SlotsRemaining int `json:"slots_remaining"`
	}
	_ = json.Unmarshal(body, &info)
	if info.Limit != 3 {
		t.Errorf("limit = %d, want 3", info.Limit)
	}
	if info.TeamsCount < 1 {
		t.Errorf("teams_count = %d, want ≥1", info.TeamsCount)
	}
	if info.SlotsRemaining < 0 {
		t.Errorf("slots_remaining negative: %d", info.SlotsRemaining)
	}
}

func TestAdminSetSubscription_PaidPlan(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "subadmin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "subadmin@example.com")

	owner := createUser(t, pool, "subOwner@example.com")
	team := createTeamDirect(t, pool, "Sub", owner)

	status, body := do(t, app, "PATCH", "/v1/admin/teams/"+team.String()+"/subscription", tok, map[string]any{
		"plan":  "team",
		"until": "2027-12-31T23:59:59Z",
		"note":  "Y combinator deal",
		"payment": map[string]any{
			"amount_cents": 50000,
			"currency":     "USD",
			"method":       "manual_transfer",
			"note":         "wire-2027-01",
			"covers_until": "2027-12-31T23:59:59Z",
		},
	})
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}

	var plan, note string
	_ = pool.QueryRow(context.Background(),
		"SELECT subscription_plan, subscription_note FROM teams WHERE id=$1", team).Scan(&plan, &note)
	if plan != "team" {
		t.Errorf("plan = %q, want team", plan)
	}
	if note != "Y combinator deal" {
		t.Errorf("note = %q", note)
	}

	var paymentCount int
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM team_payments WHERE team_id=$1", team).Scan(&paymentCount)
	if paymentCount != 1 {
		t.Errorf("payment count = %d, want 1", paymentCount)
	}
}

func TestAdminListPayments(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "payadmin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "payadmin@example.com")

	owner := createUser(t, pool, "payOwner@example.com")
	team := createTeamDirect(t, pool, "Pay", owner)

	_, err := pool.Exec(context.Background(),
		`INSERT INTO team_payments (team_id, amount_cents, currency, method, covers_until, recorded_by)
		 VALUES ($1, 10000, 'USD', 'cash', NOW() + INTERVAL '1 year', $2)`,
		team, admin)
	if err != nil {
		t.Fatalf("insert payment: %v", err)
	}

	status, body := do(t, app, "GET", "/v1/admin/teams/"+team.String()+"/payments", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var out struct {
		Payments []map[string]any `json:"payments"`
	}
	_ = json.Unmarshal(body, &out)
	if len(out.Payments) != 1 {
		t.Errorf("payments len = %d, want 1", len(out.Payments))
	}
}

func TestAdminDeleteTeam(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "deladmin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "deladmin@example.com")

	owner := createUser(t, pool, "delOwner@example.com")
	team := createTeamDirect(t, pool, "DelMe", owner)

	status, _ := do(t, app, "DELETE", "/v1/admin/teams/"+team.String(), tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}

	var count int
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM teams WHERE id=$1", team).Scan(&count)
	if count != 0 {
		t.Error("team не удалена")
	}
}

func TestAdminAddMember(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	admin := createUser(t, pool, "admin6@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin6@example.com")

	owner := createUser(t, pool, "ownerAdd@example.com")
	team := createTeamDirect(t, pool, "Add", owner)
	target := createUser(t, pool, "addable@example.com")
	_ = target

	status, _ := do(t, app, "POST", "/v1/admin/teams/"+team.String()+"/members", tok, map[string]string{
		"email": "addable@example.com",
		"role":  "member",
	})
	if status != 200 {
		t.Fatalf("status=%d", status)
	}

	var role string
	_ = pool.QueryRow(context.Background(),
		"SELECT role FROM team_members WHERE team_id=$1 AND user_id=$2", team, target).Scan(&role)
	if role != "member" {
		t.Errorf("role = %q, want member", role)
	}
}
