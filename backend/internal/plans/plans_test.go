package plans

import "testing"

func TestFor_KnownPlans(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"free", PlanFree},
		{"pro", PlanPro},
		{"business", PlanBusiness},
		{"enterprise", PlanEnterprise},
		{"PRO", PlanPro},        // case-insensitive
		{"  business  ", PlanBusiness}, // trim
		{"", PlanFree},          // empty → free
		{"garbage", PlanFree},   // unknown → free
	}
	for _, tc := range cases {
		got := For(tc.in)
		if got.Plan != tc.want {
			t.Errorf("For(%q).Plan = %q, want %q", tc.in, got.Plan, tc.want)
		}
	}
}

func TestPresets_Shape(t *testing.T) {
	if Free.MaxUsersPerTeam != 5 {
		t.Errorf("Free.MaxUsersPerTeam = %d, want 5 (pricing page contract)", Free.MaxUsersPerTeam)
	}
	if Free.MaxWebhooks != 1 {
		t.Errorf("Free.MaxWebhooks = %d, want 1", Free.MaxWebhooks)
	}
	if Free.RetentionDays != 30 {
		t.Errorf("Free.RetentionDays = %d, want 30", Free.RetentionDays)
	}
	if Free.SSO {
		t.Error("Free.SSO should be false")
	}
	if Pro.MaxUsersPerTeam != 50 {
		t.Errorf("Pro.MaxUsersPerTeam = %d, want 50", Pro.MaxUsersPerTeam)
	}
	if !Business.SSO || !Business.AuditLog || !Business.CustomRoles {
		t.Error("Business should have SSO, AuditLog, CustomRoles enabled")
	}
	if Enterprise.MaxUsersPerTeam != 0 {
		t.Errorf("Enterprise.MaxUsersPerTeam = %d, want 0 (unlimited)", Enterprise.MaxUsersPerTeam)
	}
}

func TestService_EnforceOff_ReturnsUnlimited(t *testing.T) {
	s := Service{Enforce: false}
	l := s.Limits("free")
	if l.MaxUsersPerTeam != 0 {
		t.Errorf("MaxUsersPerTeam = %d, want 0 when Enforce=false", l.MaxUsersPerTeam)
	}
	if !l.SSO {
		t.Error("SSO should be on when Enforce=false (beta — all features free)")
	}
	if l.Plan != "free" {
		t.Errorf("Plan label = %q, want 'free' (preserve original tier for telemetry)", l.Plan)
	}
}

func TestService_EnforceOn_AppliesPreset(t *testing.T) {
	s := Service{Enforce: true}
	l := s.Limits("free")
	if l.MaxUsersPerTeam != 5 {
		t.Errorf("MaxUsersPerTeam = %d, want 5 when Enforce=true", l.MaxUsersPerTeam)
	}
	if l.SSO {
		t.Error("Free.SSO should be off when Enforce=true")
	}
	if s.Limits("business").SSO != true {
		t.Error("Business.SSO should be on when Enforce=true")
	}
}

func TestService_EnforceOff_EmptyPlan_DefaultsToFreeLabel(t *testing.T) {
	s := Service{Enforce: false}
	l := s.Limits("")
	if l.Plan != "free" {
		t.Errorf("empty plan → Plan label = %q, want 'free'", l.Plan)
	}
}
