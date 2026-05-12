export function formatShortDisplayName(
  first?: string | null,
  last?: string | null,
  fallback?: string,
): string {
  const f = first?.trim() ?? "";
  const l = last?.trim() ?? "";
  if (f && l) {
    const initial = l[0]?.toLocaleUpperCase() ?? "";
    return `${f} ${initial}.`;
  }
  if (f) return f;
  if (l) return l;
  return fallback ?? "—";
}
