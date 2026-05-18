package adminapp

import (
	"net/mail"
	"strings"
)

const maxDisplayNameLen = 64

func normalizeEmail(s string) (string, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || len(s) > 254 {
		return "", false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return "", false
	}
	return addr.Address, true
}

func normalizeDisplayName(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > maxDisplayNameLen {
		return "", false
	}
	if strings.ContainsAny(s, "\r\n\t") {
		return "", false
	}
	return s, true
}

func normalizeMemberRole(role string) (string, bool) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return "member", true
	}
	if role != "owner" && role != "admin" && role != "member" {
		return "", false
	}
	return role, true
}

func normalizeGlobalRole(role string) (string, bool) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "user" && role != "super_admin" {
		return "", false
	}
	return role, true
}

func normalizePlan(plan string) (string, bool) {
	plan = strings.ToLower(strings.TrimSpace(plan))
	switch plan {
	case "free", "pro", "team", "enterprise":
		return plan, true
	default:
		return "", false
	}
}
