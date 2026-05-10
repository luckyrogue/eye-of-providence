//go:build integration

package teams

import (
	"context"
	"encoding/json"
	"testing"
)

// --- POST /v1/teams/:id/invites ---

func TestCreateInvite_LinkOnly(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-inv@example.com")
	team := createTeamDirect(t, pool, "Inv", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-inv@example.com")

	status, body := do(t, app, "POST", "/v1/teams/"+team.String()+"/invites", tok, map[string]any{})
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}
	var out struct {
		Code    string `json:"code"`
		MaxUses int    `json:"max_uses"`
	}
	_ = json.Unmarshal(body, &out)
	if out.Code == "" {
		t.Error("empty invite code")
	}
	if out.MaxUses != 10 {
		t.Errorf("max_uses=%d, want 10 (link-only)", out.MaxUses)
	}
}

func TestCreateInvite_WithEmail_OneUse(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-eml@example.com")
	team := createTeamDirect(t, pool, "Eml", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-eml@example.com")

	status, body := do(t, app, "POST", "/v1/teams/"+team.String()+"/invites", tok, map[string]any{
		"email": "newbie@example.com",
	})
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}
	var out struct {
		MaxUses int    `json:"max_uses"`
		Email   string `json:"email"`
	}
	_ = json.Unmarshal(body, &out)
	if out.MaxUses != 1 {
		t.Errorf("max_uses=%d, want 1 (email-invite)", out.MaxUses)
	}
	if out.Email != "newbie@example.com" {
		t.Errorf("email=%q, want newbie@example.com", out.Email)
	}
}

func TestCreateInvite_BadEmail_400(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-bad-eml@example.com")
	team := createTeamDirect(t, pool, "BadEml", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-bad-eml@example.com")

	status, _ := do(t, app, "POST", "/v1/teams/"+team.String()+"/invites", tok, map[string]any{
		"email": "not-an-email",
	})
	if status != 400 {
		t.Errorf("status=%d, want 400", status)
	}
}

func TestCreateInvite_MemberCannot(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-mci@example.com")
	team := createTeamDirect(t, pool, "MCI", ownerID)
	memberID := createUser(t, pool, "mem-ci@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)
	tok := loginAs(t, pool, svc.JWTSecret, memberID, "mem-ci@example.com")

	status, _ := do(t, app, "POST", "/v1/teams/"+team.String()+"/invites", tok, map[string]any{})
	if status != 403 {
		t.Errorf("status=%d, want 403", status)
	}
}

// --- GET /v1/invites/:code (preview, no auth) ---

func TestInvitePreview_HappyPath(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-prev@example.com")
	team := createTeamDirect(t, pool, "PreviewTeam", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-prev@example.com")

	_, body := do(t, app, "POST", "/v1/teams/"+team.String()+"/invites", tok, map[string]any{})
	var inv struct{ Code string }
	_ = json.Unmarshal(body, &inv)

	status, body := do(t, app, "GET", "/v1/invites/"+inv.Code, "", nil)
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}
	var out struct {
		Valid    bool   `json:"valid"`
		TeamName string `json:"team_name"`
	}
	_ = json.Unmarshal(body, &out)
	if !out.Valid || out.TeamName != "PreviewTeam" {
		t.Errorf("preview = %+v", out)
	}
}

func TestInvitePreview_BadCode_404(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)

	status, _ := do(t, app, "GET", "/v1/invites/nonsense", "", nil)
	if status != 404 {
		t.Errorf("status=%d, want 404", status)
	}
}

// --- POST /v1/invites/:code/accept ---

func TestInviteAccept_HappyPath(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-acc@example.com")
	team := createTeamDirect(t, pool, "AccTeam", ownerID)
	ownerTok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-acc@example.com")
	_, inviteBody := do(t, app, "POST", "/v1/teams/"+team.String()+"/invites", ownerTok, map[string]any{})
	var inv struct{ Code string }
	_ = json.Unmarshal(inviteBody, &inv)

	guestID := createUser(t, pool, "guest-acc@example.com")
	guestTok := loginAs(t, pool, svc.JWTSecret, guestID, "guest-acc@example.com")

	status, _ := do(t, app, "POST", "/v1/invites/"+inv.Code+"/accept", guestTok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var role string
	_ = pool.QueryRow(context.Background(),
		"SELECT role FROM team_members WHERE team_id=$1 AND user_id=$2", team, guestID).Scan(&role)
	if role != "member" {
		t.Errorf("role=%q, want member", role)
	}
}

func TestInviteAccept_OneUseExhausted(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-exh@example.com")
	team := createTeamDirect(t, pool, "Exh", ownerID)
	ownerTok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-exh@example.com")
	_, inviteBody := do(t, app, "POST", "/v1/teams/"+team.String()+"/invites", ownerTok, map[string]any{
		"email": "first@example.com",
	})
	var inv struct{ Code string }
	_ = json.Unmarshal(inviteBody, &inv)

	guest1 := createUser(t, pool, "g1-exh@example.com")
	tok1 := loginAs(t, pool, svc.JWTSecret, guest1, "g1-exh@example.com")
	if status, _ := do(t, app, "POST", "/v1/invites/"+inv.Code+"/accept", tok1, nil); status != 200 {
		t.Fatalf("first accept status=%d", status)
	}

	guest2 := createUser(t, pool, "g2-exh@example.com")
	tok2 := loginAs(t, pool, svc.JWTSecret, guest2, "g2-exh@example.com")
	status, _ := do(t, app, "POST", "/v1/invites/"+inv.Code+"/accept", tok2, nil)
	if status != 404 {
		t.Errorf("second accept status=%d, want 404 (use exhausted)", status)
	}
}

// --- /v1/teams (list) and /v1/beta/info ---

func TestListMyTeams_OnlyMyTeams(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	meID := createUser(t, pool, "me-list@example.com")
	otherID := createUser(t, pool, "other-list@example.com")
	myTeam := createTeamDirect(t, pool, "Mine", meID)
	_ = createTeamDirect(t, pool, "Theirs", otherID)
	tok := loginAs(t, pool, svc.JWTSecret, meID, "me-list@example.com")

	status, body := do(t, app, "GET", "/v1/teams", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var out struct {
		Teams []map[string]any `json:"teams"`
	}
	_ = json.Unmarshal(body, &out)
	if len(out.Teams) != 1 {
		t.Fatalf("teams=%d, want 1", len(out.Teams))
	}
	if id, _ := out.Teams[0]["id"].(string); id != myTeam.String() {
		t.Errorf("got team id=%v, want %v", id, myTeam)
	}
}

// --- /v1/teams/:id (detail) ---

func TestTeamDetail_MemberCanRead(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-det@example.com")
	team := createTeamDirect(t, pool, "DetailTeam", ownerID)
	memberID := createUser(t, pool, "mem-det@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)
	tok := loginAs(t, pool, svc.JWTSecret, memberID, "mem-det@example.com")

	status, body := do(t, app, "GET", "/v1/teams/"+team.String()+"/", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var out struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}
	_ = json.Unmarshal(body, &out)
	if out.Name != "DetailTeam" || out.Role != "member" {
		t.Errorf("got %+v", out)
	}
}

func TestTeamDetail_NonMember_403(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-d2@example.com")
	team := createTeamDirect(t, pool, "D2", ownerID)
	otherID := createUser(t, pool, "other-d2@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, otherID, "other-d2@example.com")

	status, _ := do(t, app, "GET", "/v1/teams/"+team.String()+"/", tok, nil)
	if status != 403 {
		t.Errorf("status=%d, want 403", status)
	}
}
