export const PROVENANCE_BUCKETS = [
  { key: "ai_inline", labelKey: "dashboard.provenance_ai_inline", color: "hsl(var(--accent))" },
  { key: "paste_ai", labelKey: "dashboard.provenance_paste_ai", color: "#60a5fa" },
  { key: "manual", labelKey: "dashboard.provenance_manual", color: "#4ade80" },
  { key: "ai_agent", labelKey: "dashboard.provenance_ai_agent", color: "#c084fc" },
  { key: "unknown", labelKey: "dashboard.provenance_unknown", color: "rgba(255,255,255,0.18)" },
] as const;
