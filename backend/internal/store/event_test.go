package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestEventCrossPlatformContract(t *testing.T) {
	e := Event{
		TS:           time.Date(2026, 5, 4, 16, 0, 0, 0, time.UTC),
		UserID:       "12345678-1234-1234-1234-123456789012",
		DeviceID:     "00000000-0000-0000-0000-000000000001",
		AppBundle:    "com.microsoft.VSCode",
		Category:     "manual",
		Source:       "os",
		FileLang:     "rust",
		DurationMS:   30000,
		CharsIn:      120,
		LinesAdded:   5,
		LinesRemoved: 2,
	}

	got, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"ts":"2026-05-04T16:00:00Z","user_id":"12345678-1234-1234-1234-123456789012","device_id":"00000000-0000-0000-0000-000000000001","session_id":"","app_bundle":"com.microsoft.VSCode","category":"manual","source":"os","file_lang":"rust","duration_ms":30000,"chars_in":120,"lines_added":5,"lines_removed":2}`
	if string(got) != want {
		t.Errorf("contract drift detected\n got: %s\nwant: %s", got, want)
	}
}

func TestEventMetaRoundtrip(t *testing.T) {
	// agent отправляет meta как JSON-объект; backend десериализует и
	// ClickHouse-инсерт пакует обратно в String. Тут проверяем что
	// JSON-контракт стабилен: ключи отсортированы (BTreeMap у Rust-агента),
	// пустые карты не сериализуются.
	src := `{"ts":"2026-05-04T16:00:00Z","user_id":"00000000-0000-0000-0000-000000000001","device_id":"","session_id":"","app_bundle":"firefox","category":"other","source":"os","duration_ms":0,"chars_in":0,"lines_added":0,"lines_removed":0,"meta":{"clipboard_bytes":"42","clipboard_sha256":"abc"}}`
	var e Event
	if err := json.Unmarshal([]byte(src), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.Meta["clipboard_sha256"] != "abc" || e.Meta["clipboard_bytes"] != "42" {
		t.Fatalf("meta lost in roundtrip: %#v", e.Meta)
	}
	out, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != src {
		t.Errorf("meta JSON drift\n got: %s\nwant: %s", out, src)
	}
}

func TestEventStoreInsertContract(t *testing.T) {
	ctx := context.Background()
	mem := NewMemory()

	now := time.Now().UTC()
	uid := "12345678-1234-1234-1234-123456789012"
	events := []Event{
		{TS: now, UserID: uid, AppBundle: "vscode", Category: "manual", Source: "os", DurationMS: 1000},
		{TS: now, UserID: uid, AppBundle: "chatgpt.com", Category: "ai", Source: "browser", AIProvider: "openai", DurationMS: 2000},
	}

	if err := mem.Insert(ctx, events); err != nil {
		t.Fatal(err)
	}
	got, err := mem.ListRecent(ctx, uid, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Category == got[1].Category {
		t.Errorf("expected different categories preserved, got both %q", got[0].Category)
	}
}

// MemoryStore не считает provenance: её вычисляет attribution worker в
// ClickHouse. Контракт — пустая карта без ошибки, чтобы дашборд без CH
// показывал пустой донат, а не 500.
func TestMemoryStoreAggregateProvenanceIsEmpty(t *testing.T) {
	ctx := context.Background()
	mem := NewMemory()

	now := time.Now().UTC()
	uid := "12345678-1234-1234-1234-123456789012"
	if err := mem.Insert(ctx, []Event{
		{TS: now, UserID: uid, AppBundle: "vscode", Category: "manual", Source: "ide", DurationMS: 1000},
	}); err != nil {
		t.Fatal(err)
	}

	agg, err := mem.AggregateProvenance(ctx, uid, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if agg == nil {
		t.Fatal("expected empty map, got nil — вызывающий код итерирует результат")
	}
	if len(agg) != 0 {
		t.Errorf("expected empty provenance, got %v", agg)
	}
}
