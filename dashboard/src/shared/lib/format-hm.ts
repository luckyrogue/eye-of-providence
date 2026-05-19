/** Миллисекунды → KPI-формат "Xh Ymin" (value + unit). */
export function formatHm(ms: number): { value: string; unit: string } {
  const totalMin = Math.round(ms / 60000);
  const h = Math.floor(totalMin / 60);
  const m = totalMin % 60;
  if (h > 0) return { value: String(h), unit: `h ${String(m).padStart(2, "0")}m` };
  return { value: String(m), unit: "min" };
}
