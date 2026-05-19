import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useRecent } from "@/entities/event";
import { formatLogRow } from "../lib/format-log-row";

export function AttributionLog() {
  const { t } = useTranslation("app");
  const recent = useRecent(20);
  const rows = useMemo(() => (recent.data ?? []).map(formatLogRow), [recent.data]);

  return (
    <div className="eop-card col-span-12">
      <div className="card-head">
        <div>
          <div className="card-title">{t("dashboard.log_title")}</div>
          <div className="card-sub">{t("dashboard.log_sub")}</div>
        </div>
        <span
          className="inline-flex items-center gap-1.5 font-mono text-[11px]"
          style={{ color: "hsl(var(--success))" }}
        >
          <span
            className="w-1.5 h-1.5 rounded-full"
            style={{
              background: "hsl(var(--success))",
              boxShadow: "0 0 6px hsl(var(--success))",
              animation: "eop-pulse 1.6s ease-in-out infinite",
            }}
          />
          LIVE
        </span>
      </div>
      <div className="log-table">
        <div className="log-row head">
          <span>{t("dashboard.log_col_time")}</span>
          <span>{t("dashboard.log_col_tag")}</span>
          <span>{t("dashboard.log_col_file")}</span>
          <span>Δ</span>
          <span />
        </div>
        {rows.length === 0 ? (
          <div className="text-[12px] text-muted-foreground py-3">{t("dashboard.no_data_yet")}</div>
        ) : (
          rows.map((r, i) => (
            <div key={i} className="log-row">
              <span className="log-time">{r.ts}</span>
              <span className={`tag ${r.tag}`}>{r.tag}</span>
              <span className="log-file truncate">
                {r.file} <span className="path">· {r.src}</span>
              </span>
              <span className="log-lines">{r.lines}</span>
              <span />
            </div>
          ))
        )}
      </div>
    </div>
  );
}
