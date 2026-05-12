// flags.go — per-team feature flags + plan-limit overrides validation.
//
// Используется admin endpoints (`PATCH /v1/admin/teams/:id/flags` и
// `PATCH /v1/admin/teams/:id/plan-limits`) для строгого allowlist'инга
// произвольного JSONB payload'а от super_admin'а.
//
// Дизайн:
//   * Один центральный source of truth для известных ключей; добавление
//     нового flag/override ключа требует расширения этого файла + ревью
//     с product/QA (см. .team/product-acceptance/admin-*.md).
//   * `MergeFlags` — shallow merge с поддержкой `nil`-значений (sentinel
//     "сбросить ключ"). Это match'ит UI семантику: пустая ячейка ⇒
//     возврат к plan default.
//   * Числовые ranges проверяются жёстко (return error на 0 для
//     `k_anon_threshold` и т.п.).
//
// Spec refs:
//   * `.team/product-acceptance/admin-team-flags.md` — bool/int flags.
//   * `.team/product-acceptance/admin-plan-overrides.md` — numeric overrides.

package plans

import (
	"errors"
	"fmt"
)

// FlagKey — типизированные имена feature-flag ключей (boolean / numeric
// tuning gates).
const (
	FlagEnableInsights         = "enable_insights"
	FlagEnableTeamReports      = "enable_team_reports"
	FlagEnableAnomalyDetection = "enable_anomaly_detection"
	FlagEnableWebhooks         = "enable_webhooks"
	FlagKAnonThreshold         = "k_anon_threshold"
	FlagAuditLogRetentionDays  = "audit_log_retention_days"
)

// flagSpec — внутренняя спека одного flag-ключа. type='bool'|'int'; min/max
// для int-флагов (0/0 = без ограничений).
type flagSpec struct {
	typ string
	min int
	max int
}

// flagAllowlist — единый allowlist + per-key валидация. Adding a new flag:
// добавь константу + entry + sync .team/product-acceptance/admin-team-flags.md.
var flagAllowlist = map[string]flagSpec{
	FlagEnableInsights:         {typ: "bool"},
	FlagEnableTeamReports:      {typ: "bool"},
	FlagEnableAnomalyDetection: {typ: "bool"},
	FlagEnableWebhooks:         {typ: "bool"},
	// k-anonymity threshold — нижняя граница 1 (0 уничтожает privacy guard),
	// верхняя 50 как sanity-cap (выше = отчёты практически пустые).
	FlagKAnonThreshold: {typ: "int", min: 1, max: 50},
	// audit retention min 30 days (compliance floor), max 3650 (~10 лет).
	FlagAuditLogRetentionDays: {typ: "int", min: 30, max: 3650},
}

// FlagError — typed validation error для распознавания "unknown key" vs
// "wrong type" vs "out of range". Handlers maps на httperr.* codes.
type FlagError struct {
	Code   string // "unknown_flag" | "invalid_flag_type" | "value_below_minimum" | "value_above_maximum"
	Field  string
	Detail string
	// для type-mismatch и range-violation
	Expected string
	Got      string
	Min      int
	Max      int
}

func (e *FlagError) Error() string { return e.Detail }

// ValidateFlag — проверяет одну пару (key, value). value=nil считается
// "clear override, fallback to plan default" — passes без error для
// известных ключей.
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

// ValidateFlags — batch validation. Возвращает ошибку на первой найденной
// проблеме (handler контракт: all-or-nothing, см.
// admin-team-flags.md Scenario 5). Дополнительно нормализует int-coercion
// (JSON float64 → int) — возвращает новую map'у со значениями того же типа,
// которые safely Marshal'ятся обратно.
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

// MergeFlags — shallow merge patch'а в existing. nil-значение в patch'е
// удаляет ключ из result'а (semantics: "вернуться к plan default").
// Existing никогда не мутируется — возвращается новая мапа.
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

// FlagsDiff — для audit log payload'а. Возвращает {<key>: {old, new}} по
// каждому изменённому ключу (включая "сбросить" — old=value, new=nil).
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

// --- Plan-limit overrides ---

// OverrideKey — ключи numeric override'ов. Mirror'ят plans.Limits fields,
// но allowlist отдельный, чтобы fronend не мог через PATCH /plan-limits
// случайно поменять `SSO` boolean (только flags-канал ответственен за bool
// feature gates).
const (
	OverrideMaxUsersPerTeam      = "max_users_per_team"
	OverrideMaxWebhooks          = "max_webhooks"
	OverrideMaxAPITokens         = "max_api_tokens"
	OverrideEventHistoryDays     = "event_history_days"
	OverrideMaxTeamsPerUser      = "max_teams_per_user"
	OverrideAuditLogRetention    = "audit_log_retention_days"
)

// overrideSpec — числовая спека с inclusive ranges.
type overrideSpec struct {
	min int
	max int
}

// overrideAllowlist — ranges per
// .team/product-acceptance/admin-plan-overrides.md.
var overrideAllowlist = map[string]overrideSpec{
	OverrideMaxUsersPerTeam:   {min: 1, max: 10000},
	OverrideMaxWebhooks:       {min: 0, max: 1000},
	OverrideMaxAPITokens:      {min: 0, max: 500},
	OverrideEventHistoryDays:  {min: 7, max: 3650},
	OverrideMaxTeamsPerUser:   {min: 1, max: 100},
	OverrideAuditLogRetention: {min: 30, max: 3650},
}

// ValidateOverride — один (key, value). value=nil = clear (вернуться к
// plan default).
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

// ValidateOverrides — batch validation + normalization для PATCH body.
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

// MergeOverrides — same semantics as MergeFlags. patch=nil считается
// "full reset to plan defaults" (handler передаёт `nil` в этом случае; см.
// `PATCH .../plan-limits` body `{"limits": null}`).
func MergeOverrides(existing, patch map[string]any) map[string]any {
	if patch == nil {
		return nil // sentinel: означает "удалить override row, fallback to plan"
	}
	return MergeFlags(existing, patch)
}

// --- Internal helpers ---

// coerceInt — JSON unmarshal в map[string]any даёт float64 для чисел.
// Принимаем также int / int64. Не принимаем строки — frontend обязан
// слать number.
func coerceInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case float64:
		// Защита от floating-point — требуем integer value.
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

// jsonTypeName — для error message'а ("expected bool, got string").
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

// valuesEqual — equality check для primitive JSON values (bool / int /
// float64 / string). Для int vs float64 (=int): сравниваем как int.
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

// ErrInvalidFlag — sentinel для callers, которые хотят проверить
// "вернулась ли ошибка валидации" без типизации (e.g. logging).
var ErrInvalidFlag = errors.New("invalid flag")
