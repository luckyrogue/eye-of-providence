package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultBackend = "https://eop.rysdavletov.org/api"
	timeout        = 5 * time.Second
)

type hookInput struct {
	HookEventName string         `json:"hook_event_name"`
	ToolName      string         `json:"tool_name"`
	ToolInput     map[string]any `json:"tool_input"`
	ToolResponse  map[string]any `json:"tool_response"`
}

type eventOut struct {
	AppBundle    string `json:"app_bundle"`
	Source       string `json:"source"`
	Category     string `json:"category"`
	AIProvider   string `json:"ai_provider"`
	AIChannel    string `json:"ai_channel"`
	FileLang     string `json:"file_lang,omitempty"`
	DurationMS   uint32 `json:"duration_ms"`
	CharsIn      uint32 `json:"chars_in"`
	LinesAdded   uint32 `json:"lines_added"`
	LinesRemoved uint32 `json:"lines_removed"`
}

func main() {
	token := os.Getenv("EOP_TOKEN")
	if token == "" {

		os.Exit(0)
	}
	backend := os.Getenv("EOP_BACKEND")
	if backend == "" {
		backend = defaultBackend
	}
	backend = strings.TrimRight(backend, "/")

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "eop-hook: read stdin: %v\n", err)
		os.Exit(0)
	}
	var in hookInput
	if err := json.Unmarshal(raw, &in); err != nil {

		os.Exit(0)
	}

	ev, ok := buildEvent(in)
	if !ok {
		os.Exit(0)
	}

	if err := sendEvent(backend, token, ev); err != nil {
		fmt.Fprintf(os.Stderr, "eop-hook: send: %v\n", err)
	}
}

func buildEvent(in hookInput) (eventOut, bool) {
	ev := eventOut{
		AppBundle:  "claude-code",
		Source:     "cli",
		Category:   "ai",
		AIProvider: "claude-code",
		AIChannel:  "agent",
	}

	switch in.ToolName {
	case "Write":
		content, _ := in.ToolInput["content"].(string)
		if content == "" {
			return ev, false
		}
		ev.CharsIn = clampU32(len(content))
		ev.LinesAdded = clampU32(strings.Count(content, "\n") + 1)
		path, _ := in.ToolInput["file_path"].(string)
		ev.FileLang = langFromPath(path)
	case "Edit":
		newStr, _ := in.ToolInput["new_string"].(string)
		oldStr, _ := in.ToolInput["old_string"].(string)
		if newStr == "" && oldStr == "" {
			return ev, false
		}
		ev.CharsIn = clampU32(len(newStr))
		ev.LinesAdded = clampU32(strings.Count(newStr, "\n"))
		ev.LinesRemoved = clampU32(strings.Count(oldStr, "\n"))
		path, _ := in.ToolInput["file_path"].(string)
		ev.FileLang = langFromPath(path)
	case "MultiEdit":
		edits, _ := in.ToolInput["edits"].([]any)
		var addedLines, removedLines, totalChars int
		for _, raw := range edits {
			ed, _ := raw.(map[string]any)
			n, _ := ed["new_string"].(string)
			o, _ := ed["old_string"].(string)
			totalChars += len(n)
			addedLines += strings.Count(n, "\n")
			removedLines += strings.Count(o, "\n")
		}
		if totalChars == 0 {
			return ev, false
		}
		ev.CharsIn = clampU32(totalChars)
		ev.LinesAdded = clampU32(addedLines)
		ev.LinesRemoved = clampU32(removedLines)
		path, _ := in.ToolInput["file_path"].(string)
		ev.FileLang = langFromPath(path)
	default:
		return ev, false
	}

	return ev, ev.CharsIn > 0
}

func langFromPath(path string) string {
	if path == "" {
		return ""
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".py":
		return "python"
	case ".rb":
		return "ruby"
	case ".java":
		return "java"
	case ".kt":
		return "kotlin"
	case ".swift":
		return "swift"
	case ".css":
		return "css"
	case ".html":
		return "html"
	case ".json":
		return "json"
	case ".md":
		return "markdown"
	case ".sh", ".bash":
		return "shellscript"
	case ".yml", ".yaml":
		return "yaml"
	case ".sql":
		return "sql"
	}
	return ""
}

func clampU32(n int) uint32 {
	if n < 0 {
		return 0
	}
	if n > 1<<30 {
		return 1 << 30
	}
	return uint32(n)
}

func sendEvent(backend, token string, ev eventOut) error {
	body, _ := json.Marshal(map[string][]eventOut{"events": {ev}})
	req, err := http.NewRequest("POST", backend+"/v1/ingest", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: timeout}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("ingest %d", res.StatusCode)
	}
	return nil
}
