package plans

import (
	"errors"
	"fmt"
)

const (
	FlagEnableInsights         = "enable_insights"
	FlagEnableTeamReports      = "enable_team_reports"
	FlagEnableAnomalyDetection = "enable_anomaly_detection"
	FlagEnableWebhooks         = "enable_webhooks"
	FlagKAnonThreshold         = "k_anon_threshold"
	FlagAuditLogRetentionDays  = "audit_log_retention_days"
)

type flagSpec struct {
	typ string
	min int
	max int
}

var flagAllowlist = map[string]flagSpec{
	FlagEnableInsights:         {typ: "bool"},
	FlagEnableTeamReports:      {typ: "bool"},
	FlagEnableAnomalyDetection: {typ: "bool"},
	FlagEnableWebhooks:         {typ: "bool"},

	FlagKAnonThreshold: {typ: "int", min: 1, max: 50},

	FlagAuditLogRetentionDays: {typ: "int", min: 30, max: 3650},
}

type FlagError struct {
	Code   string
	Field  string
	Detail string

	Expected string
	Got      string
	Min      int
	Max      int
}

func (e *FlagError) Error() string { return e.Detail }

func ValidateFlag(key string, value any) error {
	spec, ok := flagAllowlist[key]
	if !ok {
		return &FlagError{
			Code:   "unknown_flag",
			Field:  key,
			Detail: fmt.Sprintf("unknown flag: %s", key),
		}
	}
	if value == nil {
		return nil
	}
	switch spec.typ {
	case "bool":
		if _, ok := value.(bool); !ok {
			return &FlagError{
				Code:     "invalid_flag_type",
				Field:    key,
				Expected: "bool",
				Got:      jsonTypeName(value),
				Detail:   fmt.Sprintf("%s: expected bool, got %s", key, jsonTypeName(value)),
			}
		}
	case "int":
		n, ok := coerceInt(value)
		if !ok {
			return &FlagError{
				Code:     "invalid_flag_type",
				Field:    key,
				Expected: "int",
				Got:      jsonTypeName(value),
				Detail:   fmt.Sprintf("%s: expected int, got %s", key, jsonTypeName(value)),
			}
		}
		if spec.min != 0 && n < spec.min {
			return &FlagError{
				Code:   "value_below_minimum",
				Field:  key,
				Min:    spec.min,
				Detail: fmt.Sprintf("%s: %d below minimum %d", key, n, spec.min),
			}
		}
		if spec.max != 0 && n > spec.max {
			return &FlagError{
				Code:   "value_above_maximum",
				Field:  key,
				Max:    spec.max,
				Detail: fmt.Sprintf("%s: %d above maximum %d", key, n, spec.max),
			}
		}
	}
	return nil
}

func ValidateFlags(patch map[string]any) (map[string]any, error) {
	normalized := make(map[string]any, len(patch))
	for k, v := range patch {
		if err := ValidateFlag(k, v); err != nil {
			return nil, err
		}
		spec := flagAllowlist[k]
		if v != nil && spec.typ == "int" {
			n, _ := coerceInt(v)
			normalized[k] = n
		} else {
			normalized[k] = v
		}
	}
	return normalized, nil
}

func MergeFlags(existing, patch map[string]any) map[string]any {
	out := make(map[string]any, len(existing)+len(patch))
	for k, v := range existing {
		out[k] = v
	}
	for k, v := range patch {
		if v == nil {
			delete(out, k)
			continue
		}
		out[k] = v
	}
	return out
}

func FlagsDiff(before, after map[string]any) map[string]map[string]any {
	diff := map[string]map[string]any{}
	for k, vNew := range after {
		vOld, had := before[k]
		if !had || !valuesEqual(vOld, vNew) {
			diff[k] = map[string]any{"old": vOld, "new": vNew}
		}
	}
	for k, vOld := range before {
		if _, kept := after[k]; !kept {
			diff[k] = map[string]any{"old": vOld, "new": nil}
		}
	}
	return diff
}

const (
	OverrideMaxUsersPerTeam   = "max_users_per_team"
	OverrideMaxWebhooks       = "max_webhooks"
	OverrideMaxAPITokens      = "max_api_tokens"
	OverrideEventHistoryDays  = "event_history_days"
	OverrideMaxTeamsPerUser   = "max_teams_per_user"
	OverrideAuditLogRetention = "audit_log_retention_days"
)

type overrideSpec struct {
	min int
	max int
}

var overrideAllowlist = map[string]overrideSpec{
	OverrideMaxUsersPerTeam:   {min: 1, max: 10000},
	OverrideMaxWebhooks:       {min: 0, max: 1000},
	OverrideMaxAPITokens:      {min: 0, max: 500},
	OverrideEventHistoryDays:  {min: 7, max: 3650},
	OverrideMaxTeamsPerUser:   {min: 1, max: 100},
	OverrideAuditLogRetention: {min: 30, max: 3650},
}

func ValidateOverride(key string, value any) error {
	spec, ok := overrideAllowlist[key]
	if !ok {
		return &FlagError{
			Code:   "unknown_override",
			Field:  key,
			Detail: fmt.Sprintf("unknown override: %s", key),
		}
	}
	if value == nil {
		return nil
	}
	n, ok := coerceInt(value)
	if !ok {
		return &FlagError{
			Code:     "invalid_override_type",
			Field:    key,
			Expected: "int",
			Got:      jsonTypeName(value),
			Detail:   fmt.Sprintf("%s: expected int, got %s", key, jsonTypeName(value)),
		}
	}
	if n < spec.min || n > spec.max {
		return &FlagError{
			Code:   "value_out_of_range",
			Field:  key,
			Min:    spec.min,
			Max:    spec.max,
			Detail: fmt.Sprintf("%s: %d outside [%d, %d]", key, n, spec.min, spec.max),
		}
	}
	return nil
}

func ValidateOverrides(patch map[string]any) (map[string]any, error) {
	if patch == nil {
		return nil, nil
	}
	normalized := make(map[string]any, len(patch))
	for k, v := range patch {
		if err := ValidateOverride(k, v); err != nil {
			return nil, err
		}
		if v == nil {
			normalized[k] = nil
			continue
		}
		n, _ := coerceInt(v)
		normalized[k] = n
	}
	return normalized, nil
}

func MergeOverrides(existing, patch map[string]any) map[string]any {
	if patch == nil {
		return nil
	}
	return MergeFlags(existing, patch)
}

func coerceInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case float64:

		if x != float64(int(x)) {
			return 0, false
		}
		return int(x), true
	case float32:
		f := float64(x)
		if f != float64(int(f)) {
			return 0, false
		}
		return int(f), true
	}
	return 0, false
}

func jsonTypeName(v any) string {
	switch v.(type) {
	case bool:
		return "bool"
	case string:
		return "string"
	case float64, float32, int, int32, int64:
		return "number"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	case nil:
		return "null"
	}
	return "unknown"
}

func valuesEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}
	if na, ok := coerceInt(a); ok {
		if nb, ok := coerceInt(b); ok {
			return na == nb
		}
	}
	return a == b
}

var ErrInvalidFlag = errors.New("invalid flag")
