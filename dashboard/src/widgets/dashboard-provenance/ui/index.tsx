import { useTranslation } from "react-i18next";
import { useProvenance } from "@/entities/event";
import { AI_PROVENANCE_KEYS, PROVENANCE_BUCKETS } from "../model/buckets";

export function ProvenanceDonut() {
  const { t } = useTranslation("app");
  const provenance = useProvenance(7);
  const ms = provenance.data ?? {};

  // Нулевые категории не рисуем. `pasted_ai` производится только на Phase B
  // атрибуции, и подпись «Pasted-AI · 0%» читается как «вставок из AI не
  // было» — то есть как измерение, хотя это отсутствие измерения.
  const data = PROVENANCE_BUCKETS.map((b) => ({
    ...b,
    label: t(b.labelKey),
    ms: ms[b.key] ?? 0,
  })).filter((d) => d.ms > 0);

  const total = data.reduce((acc, d) => acc + d.ms, 0);
  const pcts = data.map((d) => ({
    ...d,
    pct: total > 0 ? Math.round((d.ms / total) * 100) : 0,
  }));
  const aiTotal = pcts
    .filter((d) => AI_PROVENANCE_KEYS.has(d.key))
    .reduce((acc, d) => acc + d.pct, 0);

  const r = 60;
  const c = 2 * Math.PI * r;
  let offset = 0;
  return (
    <div className="eop-card col-span-12 min-[1181px]:col-span-5">
      <div className="card-head">
        <div>
          <div className="card-title">{t("dashboard.donut_title")}</div>
          <div className="card-sub">{t("dashboard.donut_sub", { days: 7 })}</div>
        </div>
      </div>
      {total === 0 ? (
        <div className="card-sub">{t("dashboard.donut_empty")}</div>
      ) : (
        <div className="donut-row">
          <div className="donut">
            <svg viewBox="0 0 160 160" style={{ transform: "rotate(-90deg)" }}>
              <circle
                cx="80"
                cy="80"
                r={r}
                fill="none"
                stroke="rgba(255,255,255,0.04)"
                strokeWidth="14"
              />
              {pcts.map((s, i) => {
                const dash = (s.pct / 100) * c;
                const seg = (
                  <circle
                    key={i}
                    cx="80"
                    cy="80"
                    r={r}
                    fill="none"
                    stroke={s.color}
                    strokeWidth="14"
                    strokeDasharray={`${dash} ${c - dash}`}
                    strokeDashoffset={-offset}
                    style={{ transition: "stroke-dasharray 1s ease" }}
                  />
                );
                offset += dash;
                return seg;
              })}
            </svg>
            <div className="donut-center">
              <div>
                <div className="big">{aiTotal}%</div>
                <div className="lil">{t("dashboard.donut_center")}</div>
              </div>
            </div>
          </div>
          <div className="legend-row">
            {pcts.map((p) => (
              <div key={p.label} className="item">
                <span className="sw" style={{ background: p.color }} />
                <span>{p.label}</span>
                <span className="pct">{p.pct}%</span>
                <span className="lines">{Math.round(p.ms / 1000)}s</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
