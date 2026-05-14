package prcomment

import (
	"fmt"
	"strings"
)

func FormatComment(agg Aggregate, dashboardURL string) string {
	var b strings.Builder
	b.WriteString("### :eye: Eye of Providence — coding attribution\n\n")

	if agg.TotalCommits == 0 {
		b.WriteString("_No commits matched in EoP database. Make sure agents are running and commits ingested._\n")
		return b.String()
	}

	if agg.AIPercentWeighted == nil {
		fmt.Fprintf(&b,
			"_%d commits found, but no AI-attribution data yet._ Install the EoP agent to track AI vs manual coding time:\n\n",
			agg.TotalCommits,
		)
		b.WriteString("- macOS / Windows desktop: https://eop.rysdavletov.org/downloads\n")
		b.WriteString("- VS Code marketplace: search \"Eye of Providence\"\n")
		b.WriteString("- Claude Code hook: `eop-hook` (see docs)\n")
		return b.String()
	}

	pct := *agg.AIPercentWeighted
	bar := progressBar(pct)
	fmt.Fprintf(&b, "**AI-assisted: %.0f%%** %s\n\n", pct, bar)

	b.WriteString("| Metric | Value |\n")
	b.WriteString("|---|---|\n")
	fmt.Fprintf(&b, "| Commits | %d (with attribution: %d) |\n", agg.TotalCommits, agg.WithAttribution)
	fmt.Fprintf(&b, "| Lines added | +%d |\n", agg.LinesAdded)
	fmt.Fprintf(&b, "| Lines removed | -%d |\n", agg.LinesRemoved)

	if dashboardURL != "" {
		fmt.Fprintf(&b, "\n[View team breakdown →](%s)\n", dashboardURL)
	}
	return b.String()
}

func progressBar(pct float64) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct/10 + 0.5)
	if filled > 10 {
		filled = 10
	}
	return "`" + strings.Repeat("█", filled) + strings.Repeat("░", 10-filled) + "`"
}
