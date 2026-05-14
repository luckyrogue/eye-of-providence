//go:build integration

package teams

import (
	"context"
	"encoding/json"
	"testing"
)

func TestListMembers_HappyPath(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-lm@example.com")
	team := createTeamDirect(t, pool, "MembersTest", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-lm@example.com")

	memberID := createUser(t, pool, "member-lm@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)

	status, body := do(t, app, "GET", "/v1/teams/"+team.String()+"/members", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}
	var out struct {
		Members []map[string]any `json:"members"`
	}
	_ = json.Unmarshal(body, &out)
	if len(out.Members) != 2 {
		t.Errorf("members = %d, want 2", len(out.Members))
	}
}

func TestListMembers_NotMember_403(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-out@example.com")
	team := createTeamDirect(t, pool, "Out", ownerID)
	otherID := createUser(t, pool, "outsider@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, otherID, "outsider@example.com")

	status, _ := do(t, app, "GET", "/v1/teams/"+team.String()+"/members", tok, nil)
	if status != 403 {
		t.Errorf("status=%d, want 403", status)
	}
}

func TestUpdateTeam_RenameByOwner(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-up@example.com")
	team := createTeamDirect(t, pool, "OldName", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-up@example.com")

	status, _ := do(t, app, "PATCH", "/v1/teams/"+team.String()+"/", tok, map[string]string{"name": "NewName"})
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var name string
	_ = pool.QueryRow(context.Background(), "SELECT name FROM teams WHERE id=$1", team).Scan(&name)
	if name != "NewName" {
		t.Errorf("name = %q, want NewName", name)
	}
}

func TestUpdateTeam_NonOwner_403(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-up2@example.com")
	team := createTeamDirect(t, pool, "Name", ownerID)
	memberID := createUser(t, pool, "member-up@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)
	tok := loginAs(t, pool, svc.JWTSecret, memberID, "member-up@example.com")

	status, _ := do(t, app, "PATCH", "/v1/teams/"+team.String()+"/", tok, map[string]string{"name": "Hijack"})
	if status != 403 {
		t.Errorf("status=%d, want 403", status)
	}
}

func TestDeleteTeam_OwnerCascades(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-del@example.com")
	team := createTeamDirect(t, pool, "Doomed", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-del@example.com")

	status, _ := do(t, app, "DELETE", "/v1/teams/"+team.String()+"/", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var count int
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM teams WHERE id=$1", team).Scan(&count)
	if count != 0 {
		t.Error("team не удалена")
	}

	var memberCount int
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM team_members WHERE team_id=$1", team).Scan(&memberCount)
	if memberCount != 0 {
		t.Errorf("team_members count=%d, FK cascade сломан?", memberCount)
	}
}

func TestRemoveMember_OwnerCanRemoveMember(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-rm@example.com")
	team := createTeamDirect(t, pool, "Rm", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-rm@example.com")

	memberID := createUser(t, pool, "member-rm@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)

	status, _ := do(t, app, "DELETE", "/v1/teams/"+team.String()+"/members/"+memberID.String(), tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var count int
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM team_members WHERE team_id=$1 AND user_id=$2", team, memberID).Scan(&count)
	if count != 0 {
		t.Error("member не удалён")
	}
}

func TestRemoveMember_CannotRemoveLastOwner(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-last@example.com")
	team := createTeamDirect(t, pool, "Last", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-last@example.com")

	status, _ := do(t, app, "DELETE", "/v1/teams/"+team.String()+"/members/"+ownerID.String(), tok, nil)
	if status != 409 {
		t.Errorf("status=%d, want 409 (нельзя удалить последнего owner)", status)
	}
}

func TestRemoveMember_AdminCanRemoveMember(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-arm@example.com")
	team := createTeamDirect(t, pool, "ARM", ownerID)

	adminID := createUser(t, pool, "admin-arm@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'admin')", team, adminID)
	memberID := createUser(t, pool, "member-arm@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)
	adminTok := loginAs(t, pool, svc.JWTSecret, adminID, "admin-arm@example.com")

	status, _ := do(t, app, "DELETE", "/v1/teams/"+team.String()+"/members/"+memberID.String(), adminTok, nil)
	if status != 200 {
		t.Errorf("status=%d, want 200 (admin может удалять member'а)", status)
	}
}

func TestRemoveMember_MemberCannotRemove(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-mcr@example.com")
	team := createTeamDirect(t, pool, "MCR", ownerID)

	mem1 := createUser(t, pool, "mem1@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, mem1)
	mem2 := createUser(t, pool, "mem2@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, mem2)
	tok := loginAs(t, pool, svc.JWTSecret, mem1, "mem1@example.com")

	status, _ := do(t, app, "DELETE", "/v1/teams/"+team.String()+"/members/"+mem2.String(), tok, nil)
	if status != 403 {
		t.Errorf("status=%d, want 403", status)
	}
}

func TestUpdateMemberRole_OwnerPromotesMemberToAdmin(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-promote@example.com")
	team := createTeamDirect(t, pool, "Promote", ownerID)
	memberID := createUser(t, pool, "promote-target@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-promote@example.com")

	status, _ := do(t, app, "PATCH", "/v1/teams/"+team.String()+"/members/"+memberID.String(), tok, map[string]string{"role": "admin"})
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var role string
	_ = pool.QueryRow(context.Background(),
		"SELECT role FROM team_members WHERE team_id=$1 AND user_id=$2", team, memberID).Scan(&role)
	if role != "admin" {
		t.Errorf("role=%q, want admin", role)
	}
}

func TestUpdateMemberRole_NonOwner_403(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-pno@example.com")
	team := createTeamDirect(t, pool, "Pno", ownerID)
	adminID := createUser(t, pool, "admin-pno@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'admin')", team, adminID)
	memberID := createUser(t, pool, "mem-pno@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)
	adminTok := loginAs(t, pool, svc.JWTSecret, adminID, "admin-pno@example.com")

	status, _ := do(t, app, "PATCH", "/v1/teams/"+team.String()+"/members/"+memberID.String(), adminTok, map[string]string{"role": "admin"})
	if status != 403 {
		t.Errorf("status=%d, want 403 (admin не может менять роли)", status)
	}
}

func TestUpdateMemberRole_BadRole_400(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-bad@example.com")
	team := createTeamDirect(t, pool, "Bad", ownerID)
	memberID := createUser(t, pool, "mem-bad@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-bad@example.com")

	status, _ := do(t, app, "PATCH", "/v1/teams/"+team.String()+"/members/"+memberID.String(), tok, map[string]string{"role": "god"})
	if status != 400 {
		t.Errorf("status=%d, want 400", status)
	}
}

func TestListProjects_Empty(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-pr@example.com")
	team := createTeamDirect(t, pool, "Proj", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-pr@example.com")

	status, body := do(t, app, "GET", "/v1/teams/"+team.String()+"/projects", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var out struct {
		Projects []map[string]any `json:"projects"`
	}
	_ = json.Unmarshal(body, &out)
	if len(out.Projects) != 0 {
		t.Errorf("projects = %d, want 0", len(out.Projects))
	}
}

func TestCreateProject_OwnerHappyPath(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-cp@example.com")
	team := createTeamDirect(t, pool, "CP", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-cp@example.com")

	status, _ := do(t, app, "POST", "/v1/teams/"+team.String()+"/projects", tok, map[string]string{
		"name":     "frontend",
		"repo_url": "https://github.com/eop/frontend",
	})
	if status != 200 {
		t.Fatalf("status=%d", status)
	}

	var count int
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM projects WHERE team_id=$1 AND name='frontend'", team).Scan(&count)
	if count != 1 {
		t.Errorf("project not created: count=%d", count)
	}
}

func TestCreateProject_MemberCannot(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-mc@example.com")
	team := createTeamDirect(t, pool, "MC", ownerID)
	memberID := createUser(t, pool, "member-mc@example.com")
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')", team, memberID)
	tok := loginAs(t, pool, svc.JWTSecret, memberID, "member-mc@example.com")

	status, _ := do(t, app, "POST", "/v1/teams/"+team.String()+"/projects", tok, map[string]string{
		"name": "rogue",
	})
	if status != 403 {
		t.Errorf("status=%d, want 403", status)
	}
}

func TestIngestCommit_HappyPath(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-ic@example.com")
	team := createTeamDirect(t, pool, "IC", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-ic@example.com")

	var projID string
	_ = pool.QueryRow(context.Background(),
		`INSERT INTO projects (id, user_id, team_id, name, root_path_hash)
		 VALUES (gen_random_uuid(), $1, $2, 'p', 'h') RETURNING id::text`,
		ownerID, team).Scan(&projID)

	status, _ := do(t, app, "POST", "/v1/commits", tok, map[string]any{
		"project_id":    projID,
		"sha":           "abc1234",
		"message":       "test commit",
		"branch":        "main",
		"files_changed": 2,
		"lines_added":   10,
		"lines_removed": 3,
		"authored_at":   "2026-05-09T12:00:00Z",
	})
	if status != 200 {
		t.Fatalf("status=%d", status)
	}

	var count int
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM commits WHERE sha='abc1234'").Scan(&count)
	if count != 1 {
		t.Errorf("commit count=%d, want 1", count)
	}
}

func TestIngestCommit_NonMember_403(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-icn@example.com")
	team := createTeamDirect(t, pool, "ICN", ownerID)
	var projID string
	_ = pool.QueryRow(context.Background(),
		`INSERT INTO projects (id, user_id, team_id, name, root_path_hash)
		 VALUES (gen_random_uuid(), $1, $2, 'p', 'h') RETURNING id::text`,
		ownerID, team).Scan(&projID)

	outsider := createUser(t, pool, "outsider-ic@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, outsider, "outsider-ic@example.com")

	status, _ := do(t, app, "POST", "/v1/commits", tok, map[string]any{
		"project_id":  projID,
		"sha":         "deny1234",
		"authored_at": "2026-05-09T12:00:00Z",
	})
	if status != 403 {
		t.Errorf("status=%d, want 403", status)
	}
}

func TestIngestCommit_DuplicateSHA_NoOp(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-dup@example.com")
	team := createTeamDirect(t, pool, "Dup", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-dup@example.com")

	var projID string
	_ = pool.QueryRow(context.Background(),
		`INSERT INTO projects (id, user_id, team_id, name, root_path_hash)
		 VALUES (gen_random_uuid(), $1, $2, 'p', 'h') RETURNING id::text`,
		ownerID, team).Scan(&projID)

	body := map[string]any{
		"project_id":  projID,
		"sha":         "dup1234",
		"authored_at": "2026-05-09T12:00:00Z",
	}
	do(t, app, "POST", "/v1/commits", tok, body)
	status, _ := do(t, app, "POST", "/v1/commits", tok, body)
	if status != 200 {
		t.Errorf("second insert status=%d (should be ON CONFLICT no-op)", status)
	}

	var count int
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM commits WHERE sha='dup1234'").Scan(&count)
	if count != 1 {
		t.Errorf("idempotency broken: count=%d", count)
	}
}

func TestTeamCommits_List(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-tc@example.com")
	team := createTeamDirect(t, pool, "TC", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-tc@example.com")

	var projID string
	_ = pool.QueryRow(context.Background(),
		`INSERT INTO projects (id, user_id, team_id, name, root_path_hash)
		 VALUES (gen_random_uuid(), $1, $2, 'p', 'h') RETURNING id::text`,
		ownerID, team).Scan(&projID)

	for i, sha := range []string{"sha001", "sha002", "sha003"} {
		_ = i
		_, _ = pool.Exec(context.Background(),
			`INSERT INTO commits (project_id, team_id, user_id, sha, message, authored_at)
			 VALUES ($1, $2, $3, $4, 'msg', NOW())`,
			projID, team, ownerID, sha)
	}

	status, body := do(t, app, "GET", "/v1/teams/"+team.String()+"/commits", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var out struct {
		Commits []map[string]any `json:"commits"`
	}
	_ = json.Unmarshal(body, &out)
	if len(out.Commits) != 3 {
		t.Errorf("commits = %d, want 3", len(out.Commits))
	}
}

func TestTeamSummary_NoEvents(t *testing.T) {
	pool := setupTestDB(t)
	app, svc := newTestApp(t, pool)
	ownerID := createUser(t, pool, "owner-sum@example.com")
	team := createTeamDirect(t, pool, "Sum", ownerID)
	tok := loginAs(t, pool, svc.JWTSecret, ownerID, "owner-sum@example.com")

	status, body := do(t, app, "GET", "/v1/teams/"+team.String()+"/summary", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var out struct {
		Members []map[string]any `json:"members"`
	}
	_ = json.Unmarshal(body, &out)
	if len(out.Members) != 1 {
		t.Errorf("members count = %d, want 1 (owner)", len(out.Members))
	}

	if total, ok := out.Members[0]["total_ms"].(float64); ok && total != 0 {
		t.Errorf("total_ms = %v, want 0 (no EventStore)", total)
	}
}
