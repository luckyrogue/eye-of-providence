package plans

import (
	"errors"
	"testing"
)

func TestValidateFlag_BoolOK(t *testing.T) {
	if err := ValidateFlag(FlagEnableInsights, true); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if err := ValidateFlag(FlagEnableInsights, false); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidateFlag_BoolWrongType(t *testing.T) {
	err := ValidateFlag(FlagEnableInsights, "yes")
	var fe *FlagError
	if !errors.As(err, &fe) || fe.Code != "invalid_flag_type" {
		t.Fatalf("expected FlagError invalid_flag_type, got %v", err)
	}
	if fe.Expected != "bool" || fe.Got != "string" {
		t.Errorf("unexpected expected/got: %s / %s", fe.Expected, fe.Got)
	}
}

func TestValidateFlag_UnknownKey(t *testing.T) {
	err := ValidateFlag("magic_flag", true)
	var fe *FlagError
	if !errors.As(err, &fe) || fe.Code != "unknown_flag" {
		t.Fatalf("expected unknown_flag, got %v", err)
	}
	if fe.Field != "magic_flag" {
		t.Errorf("field = %q, want magic_flag", fe.Field)
	}
}

func TestValidateFlag_IntMinMax(t *testing.T) {

	if err := ValidateFlag(FlagKAnonThreshold, 5); err != nil {
		t.Errorf("expected ok for 5: %v", err)
	}
	if err := ValidateFlag(FlagKAnonThreshold, 0); err == nil {
		t.Error("expected error for 0 (below min 1)")
	}
	if err := ValidateFlag(FlagKAnonThreshold, 51); err == nil {
		t.Error("expected error for 51 (above max 50)")
	}

	if err := ValidateFlag(FlagKAnonThreshold, float64(10)); err != nil {
		t.Errorf("expected ok for float64(10), got %v", err)
	}

	if err := ValidateFlag(FlagKAnonThreshold, 5.5); err == nil {
		t.Error("expected error for float 5.5 (not integer)")
	}
}

func TestValidateFlag_NilValueClears(t *testing.T) {
	if err := ValidateFlag(FlagEnableInsights, nil); err != nil {
		t.Errorf("nil should pass (clear semantic), got %v", err)
	}
	if err := ValidateFlag("unknown_key", nil); err == nil {
		t.Error("unknown key with nil should still error (typo protection)")
	}
}

func TestMergeFlags_ShallowMerge(t *testing.T) {
	existing := map[string]any{
		FlagEnableInsights:    false,
		FlagEnableTeamReports: true,
	}
	patch := map[string]any{
		FlagEnableInsights: true,
		FlagKAnonThreshold: 10,
	}
	out := MergeFlags(existing, patch)
	if out[FlagEnableInsights] != true {
		t.Errorf("merge: enable_insights = %v, want true", out[FlagEnableInsights])
	}
	if out[FlagEnableTeamReports] != true {
		t.Errorf("merge: enable_team_reports = %v, want true (preserved)", out[FlagEnableTeamReports])
	}
	if out[FlagKAnonThreshold] != 10 {
		t.Errorf("merge: k_anon_threshold = %v, want 10", out[FlagKAnonThreshold])
	}

	if existing[FlagEnableInsights] != false {
		t.Error("existing mutated — must be immutable")
	}
}

func TestMergeFlags_NilClearsKey(t *testing.T) {
	existing := map[string]any{
		FlagEnableInsights: false,
	}
	patch := map[string]any{
		FlagEnableInsights: nil,
	}
	out := MergeFlags(existing, patch)
	if _, present := out[FlagEnableInsights]; present {
		t.Error("nil patch value should remove the key from result")
	}
}

func TestValidateFlags_Batch_FirstErrorWins(t *testing.T) {
	_, err := ValidateFlags(map[string]any{
		FlagEnableInsights: true,
		"magic":            "drop tables",
	})
	if err == nil {
		t.Fatal("expected error on unknown_flag")
	}
}

func TestValidateFlags_Normalizes(t *testing.T) {
	norm, err := ValidateFlags(map[string]any{
		FlagKAnonThreshold: float64(15),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := norm[FlagKAnonThreshold].(int)
	if !ok || v != 15 {
		t.Errorf("normalized = %v (%T), want int 15", norm[FlagKAnonThreshold], norm[FlagKAnonThreshold])
	}
}

func TestValidateOverride_OutOfRange(t *testing.T) {
	err := ValidateOverride(OverrideMaxUsersPerTeam, 50000)
	var fe *FlagError
	if !errors.As(err, &fe) || fe.Code != "value_out_of_range" {
		t.Fatalf("expected value_out_of_range, got %v", err)
	}
	if fe.Min != 1 || fe.Max != 10000 {
		t.Errorf("range info missing: min=%d max=%d", fe.Min, fe.Max)
	}
}

func TestValidateOverride_UnknownKey(t *testing.T) {
	err := ValidateOverride("max_galaxies", 1)
	var fe *FlagError
	if !errors.As(err, &fe) || fe.Code != "unknown_override" {
		t.Fatalf("expected unknown_override, got %v", err)
	}
}

func TestValidateOverride_NilClears(t *testing.T) {
	if err := ValidateOverride(OverrideMaxUsersPerTeam, nil); err != nil {
		t.Errorf("nil should pass: %v", err)
	}
}

func TestMergeOverrides_NilPatchReturnsNil(t *testing.T) {
	out := MergeOverrides(map[string]any{"max_users_per_team": 100}, nil)
	if out != nil {
		t.Errorf("nil patch should return nil (full reset), got %v", out)
	}
}

func TestFlagsDiff_DetectsChanges(t *testing.T) {
	before := map[string]any{
		FlagEnableInsights: false,
		FlagKAnonThreshold: 5,
	}
	after := map[string]any{
		FlagEnableInsights: true,
		FlagKAnonThreshold: 5,
		FlagEnableWebhooks: true,
	}
	diff := FlagsDiff(before, after)
	if _, ok := diff[FlagEnableInsights]; !ok {
		t.Error("expected diff entry for enable_insights")
	}
	if _, ok := diff[FlagKAnonThreshold]; ok {
		t.Error("unchanged key should not appear in diff")
	}
	if _, ok := diff[FlagEnableWebhooks]; !ok {
		t.Error("added key should appear in diff")
	}
}

func TestFlagsDiff_DetectsRemoval(t *testing.T) {
	before := map[string]any{
		FlagEnableInsights: false,
	}
	after := map[string]any{}
	diff := FlagsDiff(before, after)
	e, ok := diff[FlagEnableInsights]
	if !ok {
		t.Fatal("expected removal entry")
	}
	if e["new"] != nil {
		t.Errorf("new should be nil for removal, got %v", e["new"])
	}
}
