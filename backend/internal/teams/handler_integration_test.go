//go:build integration

package teams

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/eye-of-providence/backend/internal/auth"
)

func TestCreateTeam_HappyPath(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	uid := createUser(t, pool, "owner@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, uid, "owner@example.com")

	status, body := do(t, app, "POST", "/v1/teams", tok, map[string]string{"name": "Acme"})
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}
	if jsonField(t, body, "role") != "owner" {
		t.Errorf("role = %v, want 'owner'", jsonField(t, body, "role"))
	}
	if jsonField(t, body, "name") != "Acme" {
		t.Errorf("name = %v", jsonField(t, body, "name"))
	}
}

func TestCreateTeam_OneOwnerPerUser(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	uid := createUser(t, pool, "owner@example.com")
	createTeamDirect(t, pool, "First Co", uid)
	tok := loginAs(t, pool, svc.JWTSecret, uid, "owner@example.com")

	status, body := do(t, app, "POST", "/v1/teams", tok, map[string]string{"name": "Second Co"})
	if status != 403 {
		t.Fatalf("expected 403 owner_limit, got %d body=%s", status, string(body))
	}
	if jsonField(t, body, "code") != "owner_limit" {
		t.Errorf("code = %v, want owner_limit", jsonField(t, body, "code"))
	}
}

func TestCreateTeam_BetaLimitReached(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	for i := 0; i < 3; i++ {
		other := createUser(t, pool, "u"+string(rune('0'+i))+"@example.com")
		createTeamDirect(t, pool, "Team "+string(rune('A'+i)), other)
	}
	uid := createUser(t, pool, "newcomer@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, uid, "newcomer@example.com")

	status, body := do(t, app, "POST", "/v1/teams", tok, map[string]string{"name": "Fourth"})
	if status != 403 {
		t.Fatalf("expected 403 beta_full, got %d body=%s", status, string(body))
	}
	if jsonField(t, body, "code") != "beta_full" {
		t.Errorf("code = %v, want beta_full", jsonField(t, body, "code"))
	}
}

func TestCreateTeam_SuperAdminBypassesBeta(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	for i := 0; i < 3; i++ {
		other := createUser(t, pool, "u"+string(rune('0'+i))+"@x.com")
		createTeamDirect(t, pool, "Team "+string(rune('A'+i)), other)
	}
	admin := createUser(t, pool, "admin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin@example.com")

	status, body := do(t, app, "POST", "/v1/teams", tok, map[string]string{"name": "AdminCo"})
	if status != 200 {
		t.Fatalf("super_admin should bypass beta limit, got %d body=%s", status, string(body))
	}
}

func TestUpdateRole_PromoteToOwner_RefusesIfTargetOwnsAnother(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)

	alice := createUser(t, pool, "alice@x.com")
	bob := createUser(t, pool, "bob@x.com")
	teamA := createTeamDirect(t, pool, "A", alice)
	createTeamDirect(t, pool, "B", bob)

	_, err := pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", teamA, bob)
	if err != nil {
		t.Fatal(err)
	}
	tok := loginAs(t, pool, svc.JWTSecret, alice, "alice@x.com")

	status, body := do(t, app, "PATCH", "/v1/teams/"+teamA.String()+"/members/"+bob.String(),
		tok, map[string]string{"role": "owner"})
	if status != 409 {
		t.Fatalf("expected 409 owner_limit, got %d body=%s", status, string(body))
	}
	if jsonField(t, body, "code") != "owner_limit" {
		t.Errorf("code = %v, want owner_limit", jsonField(t, body, "code"))
	}
}

func TestUpdateRole_CannotDemoteLastOwner(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	uid := createUser(t, pool, "owner@x.com")
	teamID := createTeamDirect(t, pool, "Co", uid)
	tok := loginAs(t, pool, svc.JWTSecret, uid, "owner@x.com")

	status, _ := do(t, app, "PATCH", "/v1/teams/"+teamID.String()+"/members/"+uid.String(),
		tok, map[string]string{"role": "admin"})
	if status != 409 {
		t.Errorf("expected 409 last-owner-protect, got %d", status)
	}
}

func TestSetSubscription_RollbackOnBadPayment(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	admin := createUser(t, pool, "admin@x.com")
	makeSuperAdmin(t, pool, admin)
	target := createUser(t, pool, "owner@x.com")
	teamID := createTeamDirect(t, pool, "Acme", target)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin@x.com")

	body := map[string]any{
		"plan":  "pro",
		"until": "2027-01-01T00:00:00Z",
		"payment": map[string]any{
			"amount_cents": 5000,
			"currency":     "USD",
			"method":       "manual_transfer",
			"covers_until": "BROKEN-DATE",
		},
	}
	status, _ := do(t, app, "PATCH", "/v1/admin/teams/"+teamID.String()+"/subscription", tok, body)
	if status != 400 {
		t.Fatalf("expected 400 on bad covers_until, got %d", status)
	}

	var plan string
	_ = pool.QueryRow(context.Background(),
		"SELECT subscription_plan FROM teams WHERE id=$1", teamID).Scan(&plan)
	if plan != "free" {
		t.Errorf("plan should remain 'free' on validation fail, got %q", plan)
	}
}

func TestSetSubscription_AtomicWithPayment(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	admin := createUser(t, pool, "admin@x.com")
	makeSuperAdmin(t, pool, admin)
	target := createUser(t, pool, "owner@x.com")
	teamID := createTeamDirect(t, pool, "Acme", target)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin@x.com")

	body := map[string]any{
		"plan":  "pro",
		"until": "2027-01-01T00:00:00Z",
		"payment": map[string]any{
			"amount_cents": 5000,
			"currency":     "USD",
			"method":       "manual_transfer",
			"note":         "first payment",
			"covers_until": "2027-01-01T00:00:00Z",
		},
	}
	status, _ := do(t, app, "PATCH", "/v1/admin/teams/"+teamID.String()+"/subscription", tok, body)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var plan string
	var paymentCount int
	_ = pool.QueryRow(context.Background(),
		"SELECT subscription_plan FROM teams WHERE id=$1", teamID).Scan(&plan)
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM team_payments WHERE team_id=$1", teamID).Scan(&paymentCount)
	if plan != "pro" {
		t.Errorf("plan = %q, want pro", plan)
	}
	if paymentCount != 1 {
		t.Errorf("payments = %d, want 1", paymentCount)
	}
}

func TestSetSubscription_ZeroAmountRejected(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	admin := createUser(t, pool, "admin@x.com")
	makeSuperAdmin(t, pool, admin)
	target := createUser(t, pool, "o@x.com")
	teamID := createTeamDirect(t, pool, "Acme", target)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin@x.com")

	body := map[string]any{
		"plan":  "pro",
		"until": "2027-01-01T00:00:00Z",
		"payment": map[string]any{
			"amount_cents": 0,
			"covers_until": "2027-01-01T00:00:00Z",
		},
	}
	status, _ := do(t, app, "PATCH", "/v1/admin/teams/"+teamID.String()+"/subscription", tok, body)
	if status != 400 {
		t.Errorf("expected 400 on amount=0, got %d", status)
	}
}

func TestAdminUpdateUser_DemoteBumpsTokenVersion(t *testing.T) {
	pool := setupTestDB(t)
	_, svc := newTestApp(t, pool)
	admin := createUser(t, pool, "admin@x.com")
	makeSuperAdmin(t, pool, admin)
	other := createUser(t, pool, "other@x.com")
	makeSuperAdmin(t, pool, other)

	tvBefore, _ := auth.TokenVersion(context.Background(), pool, other)

	app, _ := newTestApp(t, pool)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin@x.com")
	status, _ := do(t, app, "PATCH", "/v1/admin/users/"+other.String(),
		tok, map[string]string{"global_role": "user"})
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	tvAfter, _ := auth.TokenVersion(context.Background(), pool, other)
	if tvAfter <= tvBefore {
		t.Errorf("tv should bump on role change: before=%d after=%d", tvBefore, tvAfter)
	}
}

func TestAdminUpdateUser_CannotDemoteSelf(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	admin := createUser(t, pool, "admin@x.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin@x.com")

	status, body := do(t, app, "PATCH", "/v1/admin/users/"+admin.String(),
		tok, map[string]string{"global_role": "user"})
	if status != 409 {
		t.Fatalf("expected 409 cannot-demote-self, got %d body=%s", status, string(body))
	}
}

func TestAdminDeleteUser_CannotDeleteLastSuperAdmin(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	admin := createUser(t, pool, "admin@x.com")
	makeSuperAdmin(t, pool, admin)
	other := createUser(t, pool, "other@x.com")
	makeSuperAdmin(t, pool, other)

	other2 := createUser(t, pool, "another-other@x.com")
	makeSuperAdmin(t, pool, other2)
	_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id=$1", other2)
	_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id=$1", other)

	tok := loginAs(t, pool, svc.JWTSecret, admin, "admin@x.com")
	status, _ := do(t, app, "DELETE", "/v1/admin/users/"+admin.String(), tok, nil)
	if status != 409 {
		t.Errorf("expected 409 self/last-super, got %d", status)
	}
}

func TestMiddleware_RevokedTokenAfterTVBump(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	uid := createUser(t, pool, "user@x.com")
	createTeamDirect(t, pool, "T", uid)
	tok := loginAs(t, pool, svc.JWTSecret, uid, "user@x.com")

	status, _ := do(t, app, "GET", "/v1/teams", tok, nil)
	if status != 200 {
		t.Fatalf("expected 200 with fresh token, got %d", status)
	}

	if err := auth.BumpTokenVersion(context.Background(), pool, uid); err != nil {
		t.Fatalf("bump: %v", err)
	}
	status, body := do(t, app, "GET", "/v1/teams", tok, nil)
	if status != http.StatusUnauthorized {
		t.Errorf("expected 401 after tv bump, got %d body=%s", status, string(body))
	}
	if !strings.Contains(string(body), "session invalidated") {
		t.Errorf("expected 'session invalidated' message, got %s", string(body))
	}
}

func TestMiddleware_MissingBearer(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)
	status, _ := do(t, app, "GET", "/v1/teams", "", nil)
	if status != 401 {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestInviteAccept_ConsumesAtomically(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	owner := createUser(t, pool, "owner@x.com")
	teamID := createTeamDirect(t, pool, "T", owner)

	_, err := pool.Exec(context.Background(), `
		INSERT INTO team_invites (team_id, code, created_by, max_uses, expires_at)
		VALUES ($1, 'singleshot', $2, 1, now() + interval '1 hour')`, teamID, owner)
	if err != nil {
		t.Fatal(err)
	}

	user1 := createUser(t, pool, "u1@x.com")
	tok1 := loginAs(t, pool, svc.JWTSecret, user1, "u1@x.com")
	status, _ := do(t, app, "POST", "/v1/invites/singleshot/accept", tok1, nil)
	if status != 200 {
		t.Fatalf("first accept should succeed, got %d", status)
	}

	user2 := createUser(t, pool, "u2@x.com")
	tok2 := loginAs(t, pool, svc.JWTSecret, user2, "u2@x.com")
	status, _ = do(t, app, "POST", "/v1/invites/singleshot/accept", tok2, nil)
	if status != 404 {
		t.Errorf("second accept should fail with 404, got %d", status)
	}
}
