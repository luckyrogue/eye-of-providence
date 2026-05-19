import { Minus, TrendingDown, TrendingUp } from "lucide-react";

export function KpiTile({
  label,
  value,
  unit,
  delta,
}: {
  label: string;
  value: string;
  unit: string;
  delta?: { kind: "up" | "down" | "flat"; text: string };
}) {
  const Icon = delta?.kind === "up" ? TrendingUp : delta?.kind === "down" ? TrendingDown : Minus;
  return (
    <div className="kpi">
      <span className="kpi-label">{label}</span>
      <span className="kpi-value">
        {value}
        {unit && <span className="unit">{unit}</span>}
      </span>
      {delta && (
        <span className={`kpi-delta ${delta.kind}`}>
          <Icon className="h-3 w-3" />
          {delta.text}
        </span>
      )}
    </div>
  );
}
