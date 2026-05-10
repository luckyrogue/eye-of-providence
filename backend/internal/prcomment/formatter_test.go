package prcomment

import (
	"strings"
	"testing"
)

func TestFormatComment_Empty(t *testing.T) {
	out := FormatComment(Aggregate{}, "")
	if !strings.Contains(out, "No commits matched") {
		t.Errorf("got %q", out)
	}
}

func TestFormatComment_NoAttribution(t *testing.T) {
	out := FormatComment(Aggregate{TotalCommits: 5}, "")
	if !strings.Contains(out, "no AI-attribution data yet") {
		t.Errorf("missing graceful copy: %q", out)
	}
	if !strings.Contains(out, "Install the EoP agent") {
		t.Errorf("missing install hint: %q", out)
	}
}

func TestFormatComment_HappyPath(t *testing.T) {
	pct := 73.4
	agg := Aggregate{
		TotalCommits:      10,
		WithAttribution:   8,
		LinesAdded:        420,
		LinesRemoved:      55,
		AIPercentWeighted: &pct,
	}
	out := FormatComment(agg, "https://eop.example.com/team")
	if !strings.Contains(out, "AI-assisted: 73%") {
		t.Errorf("missing pct: %q", out)
	}
	if !strings.Contains(out, "10 (with attribution: 8)") {
		t.Errorf("missing commits row: %q", out)
	}
	if !strings.Contains(out, "+420") || !strings.Contains(out, "-55") {
		t.Errorf("missing lines: %q", out)
	}
	if !strings.Contains(out, "https://eop.example.com/team") {
		t.Errorf("missing dashboard link: %q", out)
	}
	// Progress bar — 7 filled из 10 (для 73%)
	if !strings.Contains(out, "███████░░░") {
		t.Errorf("progress bar wrong: %q", out)
	}
}

func TestFormatComment_NoDashboardURL(t *testing.T) {
	pct := 50.0
	out := FormatComment(Aggregate{TotalCommits: 1, AIPercentWeighted: &pct, LinesAdded: 10}, "")
	if strings.Contains(out, "View team breakdown") {
		t.Error("dashboard link leaked")
	}
}

func TestProgressBar(t *testing.T) {
	cases := map[float64]string{
		0:    "`░░░░░░░░░░`",
		50:   "`█████░░░░░`",
		100:  "`██████████`",
		150:  "`██████████`", // clamp to 100
		-10:  "`░░░░░░░░░░`",
		73.4: "`███████░░░`",
	}
	for in, want := range cases {
		if got := progressBar(in); got != want {
			t.Errorf("progressBar(%v) = %q, want %q", in, got, want)
		}
	}
}
