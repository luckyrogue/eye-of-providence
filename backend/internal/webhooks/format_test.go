package webhooks

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatPayload_Raw(t *testing.T) {
	body, err := formatPayload(FormatRaw, EventCommitIngested, map[string]any{"sha": "abc"})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got["event"] != EventCommitIngested {
		t.Errorf("event=%v", got["event"])
	}
	data, _ := got["data"].(map[string]any)
	if data["sha"] != "abc" {
		t.Errorf("data.sha=%v", data["sha"])
	}
	if _, ok := got["sent_at"].(string); !ok {
		t.Error("sent_at missing")
	}
}

func TestFormatPayload_DefaultFormatIsRaw(t *testing.T) {
	body, err := formatPayload("", EventCommitIngested, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"event":"commit.ingested"`) {
		t.Errorf("default not raw: %s", body)
	}
}

func TestFormatPayload_UnknownFormat(t *testing.T) {
	_, err := formatPayload("discord", EventCommitIngested, nil)
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestFormatPayload_Slack_Commit(t *testing.T) {
	body, err := formatPayload(FormatSlack, EventCommitIngested, map[string]any{
		"sha":           "abc1234567890",
		"message":       "fix: typo in README\n\nLong body here",
		"branch":        "main",
		"lines_added":   10,
		"lines_removed": 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	text, _ := got["text"].(string)
	if !strings.Contains(text, ":rocket:") {
		t.Errorf("text missing emoji: %q", text)
	}
	if !strings.Contains(text, "abc1234") {
		t.Errorf("text missing short SHA: %q", text)
	}
	if !strings.Contains(text, "fix: typo in README") {
		t.Errorf("text missing message first line: %q", text)
	}
	if strings.Contains(text, "Long body") {
		t.Errorf("text leaked message body: %q", text)
	}
	blocks, _ := got["blocks"].([]any)
	if len(blocks) < 2 {
		t.Errorf("expected >=2 blocks, got %d", len(blocks))
	}
}

func TestFormatPayload_Slack_Report(t *testing.T) {
	body, err := formatPayload(FormatSlack, EventReportGenerated, map[string]any{
		"period": "weekly_2026_W19",
		"model":  "gemini-2.5-flash",
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	text, _ := got["text"].(string)
	if !strings.Contains(text, "weekly_2026_W19") {
		t.Errorf("text missing period: %q", text)
	}
}

func TestFormatPayload_Slack_UnknownEvent(t *testing.T) {
	body, err := formatPayload(FormatSlack, "future.event", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "future.event") {
		t.Errorf("generic notification missing event name: %s", body)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("got %q", got)
	}
	if got := truncate("hellohello", 5); got != "hell…" {
		t.Errorf("got %q, want hell…", got)
	}
	// Cyrillic — runes not bytes.
	if got := truncate("привет мир", 6); got != "приве…" {
		t.Errorf("got %q", got)
	}
}

func TestFirstLine(t *testing.T) {
	if got := firstLine("a\nb"); got != "a" {
		t.Errorf("got %q", got)
	}
	if got := firstLine("single"); got != "single" {
		t.Errorf("got %q", got)
	}
}

func TestNumField(t *testing.T) {
	m := map[string]any{
		"int":    42,
		"int64":  int64(7),
		"float":  3.14,
		"string": "x",
	}
	if numField(m, "int") != 42 {
		t.Error("int")
	}
	if numField(m, "int64") != 7 {
		t.Error("int64")
	}
	if numField(m, "float") != 3 {
		t.Error("float→int truncate")
	}
	if numField(m, "string") != 0 {
		t.Error("string→0")
	}
	if numField(m, "missing") != 0 {
		t.Error("missing→0")
	}
}

func TestValidFormat(t *testing.T) {
	cases := map[string]bool{
		"raw":     true,
		"slack":   true,
		"":        false,
		"discord": false,
		"RAW":     false,
	}
	for in, want := range cases {
		if got := validFormat(in); got != want {
			t.Errorf("validFormat(%q) = %v, want %v", in, got, want)
		}
	}
}
