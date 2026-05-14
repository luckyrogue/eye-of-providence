package main

import "testing"

func TestBuildEvent_Write(t *testing.T) {
	in := hookInput{
		ToolName: "Write",
		ToolInput: map[string]any{
			"file_path": "/repo/src/foo.ts",
			"content":   "line1\nline2\nline3",
		},
	}
	ev, ok := buildEvent(in)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if ev.AIProvider != "claude-code" || ev.AIChannel != "agent" || ev.Category != "ai" {
		t.Errorf("got %+v, want claude-code/agent/ai", ev)
	}
	if ev.CharsIn != 17 {
		t.Errorf("chars=%d, want 17", ev.CharsIn)
	}
	if ev.LinesAdded != 3 {
		t.Errorf("lines_added=%d, want 3", ev.LinesAdded)
	}
	if ev.FileLang != "typescript" {
		t.Errorf("lang=%q, want typescript", ev.FileLang)
	}
}

func TestBuildEvent_Edit(t *testing.T) {
	in := hookInput{
		ToolName: "Edit",
		ToolInput: map[string]any{
			"file_path":  "/repo/main.go",
			"old_string": "a\nb",
			"new_string": "x\ny\nz\nw",
		},
	}
	ev, ok := buildEvent(in)
	if !ok {
		t.Fatal("expected ok")
	}
	if ev.LinesAdded != 3 || ev.LinesRemoved != 1 {
		t.Errorf("lines: added=%d removed=%d, want 3/1", ev.LinesAdded, ev.LinesRemoved)
	}
	if ev.FileLang != "go" {
		t.Errorf("lang=%q, want go", ev.FileLang)
	}
}

func TestBuildEvent_MultiEdit(t *testing.T) {
	in := hookInput{
		ToolName: "MultiEdit",
		ToolInput: map[string]any{
			"file_path": "/repo/x.py",
			"edits": []any{
				map[string]any{"old_string": "a", "new_string": "alpha\nbeta"},
				map[string]any{"old_string": "c", "new_string": "gamma"},
			},
		},
	}
	ev, ok := buildEvent(in)
	if !ok {
		t.Fatal("expected ok")
	}

	if ev.CharsIn != 15 {
		t.Errorf("chars=%d, want 15", ev.CharsIn)
	}
	if ev.LinesAdded != 1 {
		t.Errorf("lines_added=%d, want 1 (one \\n in 'alpha\\nbeta')", ev.LinesAdded)
	}
	if ev.FileLang != "python" {
		t.Errorf("lang=%q, want python", ev.FileLang)
	}
}

func TestBuildEvent_Read_Skipped(t *testing.T) {
	in := hookInput{ToolName: "Read", ToolInput: map[string]any{"file_path": "/foo"}}
	_, ok := buildEvent(in)
	if ok {
		t.Error("Read should not produce event")
	}
}

func TestBuildEvent_EmptyContent_Skipped(t *testing.T) {
	in := hookInput{ToolName: "Write", ToolInput: map[string]any{"file_path": "/x", "content": ""}}
	_, ok := buildEvent(in)
	if ok {
		t.Error("empty Write should not produce event")
	}
}

func TestLangFromPath(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/x/y.go", "go"},
		{"/x/y.TS", "typescript"},
		{"/x/y.tsx", "typescript"},
		{"/x/y.unknown", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := langFromPath(c.path); got != c.want {
			t.Errorf("langFromPath(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestClampU32(t *testing.T) {
	if clampU32(-1) != 0 {
		t.Error("negative clamp")
	}
	if clampU32(100) != 100 {
		t.Error("normal clamp")
	}
	max := clampU32(1<<31 - 1)
	if max != 1<<30 {
		t.Errorf("overflow clamp = %d, want 1<<30", max)
	}
}
