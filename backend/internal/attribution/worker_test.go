package attribution

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Тесты на pure classifier — он чисто-функциональный, не зависит от
// CH/PG, так что покрываем все ветки без integration setup.

func TestClassify_IDE_AI_Inline(t *testing.T) {
	got := classify(rawEvent{
		Source: "ide", Category: "ai", AIChannel: "inline", AIProvider: "copilot",
		CharsIn: 240, LinesAdded: 6,
	})
	if got.Category != "ai_inline" {
		t.Fatalf("expected ai_inline, got %q", got.Category)
	}
	if got.AIProvider != "copilot" {
		t.Fatalf("ai_provider должен сохраняться, got %q", got.AIProvider)
	}
	if got.Chars != 240 || got.Lines != 6 {
		t.Fatalf("chars/lines не пробросились: %+v", got)
	}
}

func TestClassify_CLI_Agent(t *testing.T) {
	got := classify(rawEvent{
		Source: "cli", Category: "ai", AIProvider: "anthropic", AIChannel: "agent",
		CharsIn: 0, LinesAdded: 0,
	})
	if got.Category != "ai_agent" {
		t.Fatalf("expected ai_agent, got %q", got.Category)
	}
}

func TestClassify_CLI_NoProvider_Falls_To_Unknown(t *testing.T) {
	// Без AIProvider — мы не знаем, был ли это AI-инструмент или просто
	// shell session. Консервативно → unknown, не ai_agent.
	got := classify(rawEvent{Source: "cli", Category: "manual", AIProvider: ""})
	if got.Category != "unknown" {
		t.Fatalf("expected unknown, got %q", got.Category)
	}
}

func TestClassify_IDE_Manual_Typed(t *testing.T) {
	got := classify(rawEvent{Source: "ide", Category: "manual", CharsIn: 12})
	if got.Category != "typed" {
		t.Fatalf("expected typed, got %q", got.Category)
	}
}

func TestClassify_IDE_Refactor_Preserved(t *testing.T) {
	got := classify(rawEvent{Source: "ide", Category: "refactor"})
	if got.Category != "refactor" {
		t.Fatalf("expected refactor, got %q", got.Category)
	}
}

func TestClassify_Browser_Is_Unknown_For_Now(t *testing.T) {
	// Phase A: browser-в-фокусе помечаем unknown. Phase B свяжет с
	// clipboard hash chain → pasted_ai.
	got := classify(rawEvent{Source: "browser", Category: "ai"})
	if got.Category != "unknown" {
		t.Fatalf("expected unknown (Phase A), got %q", got.Category)
	}
}

func TestClassify_OS_Focus_Unknown(t *testing.T) {
	got := classify(rawEvent{Source: "os", Category: "reading"})
	if got.Category != "unknown" {
		t.Fatalf("expected unknown, got %q", got.Category)
	}
}

func TestClassify_Preserves_Ts_And_Ids(t *testing.T) {
	ts := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	uid := uuid.New()
	pid := uuid.New()
	got := classify(rawEvent{
		TS: ts, UserID: uid, ProjectID: pid, FileLang: "rust",
		Source: "ide", Category: "manual",
		DurationMS: 7777,
	})
	if !got.TS.Equal(ts) {
		t.Fatalf("ts not preserved: %v vs %v", got.TS, ts)
	}
	if got.UserID != uid {
		t.Fatalf("user_id not preserved")
	}
	if got.ProjectID != pid {
		t.Fatalf("project_id not preserved")
	}
	if got.FileLang != "rust" {
		t.Fatalf("file_lang not preserved")
	}
	if got.FocusMS != 7777 {
		t.Fatalf("focus_ms not preserved")
	}
}

func TestClassify_AI_Channel_Required_For_Inline(t *testing.T) {
	// category='ai' БЕЗ channel='inline' — например, если IDE отметил bucket
	// неполно — НЕ должен попасть в ai_inline. Лучше undef → unknown.
	got := classify(rawEvent{Source: "ide", Category: "ai", AIChannel: ""})
	if got.Category == "ai_inline" {
		t.Fatalf("ai_inline должен требовать ai_channel='inline', got %+v", got)
	}
}
