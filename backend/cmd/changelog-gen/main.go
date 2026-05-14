package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

var publicTypes = map[string]bool{
	"feat":     true,
	"fix":      true,
	"perf":     true,
	"refactor": true,
	"docs":     true,
}

var convRe = regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?(!)?:\s*(.+)$`)

type entry struct {
	Hash     string `json:"hash"`
	Date     string `json:"date"`
	Type     string `json:"type"`
	Scope    string `json:"scope,omitempty"`
	Breaking bool   `json:"breaking,omitempty"`
	Summary  string `json:"summary"`
}

type output struct {
	GeneratedAt string  `json:"generated_at"`
	Entries     []entry `json:"entries"`
}

func main() {
	out := flag.String("out", "changelog.json", "output JSON path")
	since := flag.String("since", "", "include commits after this date (YYYY-MM-DD)")
	limit := flag.Int("limit", 200, "max commits to parse")
	flag.Parse()

	args := []string{"log", "--no-merges", "--pretty=format:%H|%aI|%s", fmt.Sprintf("-n%d", *limit)}
	if *since != "" {
		args = append(args, "--since="+*since)
	}
	cmd := exec.Command("git", args...)
	raw, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git log: %v\n", err)
		os.Exit(1)
	}

	entries := []entry{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		hash, dateStr, subject := parts[0], parts[1], parts[2]
		ts, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			continue
		}
		match := convRe.FindStringSubmatch(subject)
		if match == nil {
			continue
		}
		typ := strings.ToLower(match[1])
		if !publicTypes[typ] {
			continue
		}

		if strings.Contains(strings.ToLower(subject), "dependabot[bot]") {
			continue
		}
		entries = append(entries, entry{
			Hash:     hash[:7],
			Date:     ts.UTC().Format("2006-01-02"),
			Type:     typ,
			Scope:    match[2],
			Breaking: match[3] == "!",
			Summary:  strings.TrimSpace(match[4]),
		})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Date != entries[j].Date {
			return entries[i].Date > entries[j].Date
		}
		return entries[i].Hash > entries[j].Hash
	})

	body, err := json.MarshalIndent(output{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Entries:     entries,
	}, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, body, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %d entries to %s\n", len(entries), *out)
}
