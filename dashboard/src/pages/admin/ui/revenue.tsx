import { useTranslation } from "react-i18next";
import { useAdminRevenue, type RevenuePayment } from "../../../entities/admin";
import { formatDate } from "../../../shared/lib/tz";
export function Revenue({ tz }: { tz: string }) {
  const { t } = useTranslation("app");
  const { data, isPending, isError, error } = useAdminRevenue();
  if (isPending) {
    return (
      <div className="eop-card" style={{ minHeight: 320 }}>
        <div className="h-5 w-40 rounded mb-3" style={{ background: "hsl(var(--muted))" }} />
      </div>
    );
  }
  if (isError || !data) {
    return (
      <div
        className="eop-card"
        style={{
          border: "1px solid hsl(var(--destructive) / 0.4)",
          background: "hsl(var(--destructive) / 0.06)",
        }}
      >
        <div className="font-medium">{t("admin.error_title")}</div>
        <div className="text-[13px] text-muted-foreground mt-1">
          {error?.message ?? t("admin.error_lead")}
        </div>
      </div>
    );
  }
  return (
    <div className="space-y-4">
      <div className="kpi-grid">
        <Tile
          label={t("admin.revenue_total")}
          value={formatMoney(data.total_cents, data.currency)}
          hint={t("admin.revenue_paying_teams", { n: data.paying_teams })}
        />
        <Tile
          label={t("admin.revenue_last_30d")}
          value={formatMoney(data.last_30d_cents, data.currency)}
          hint={t("admin.revenue_mrr_proxy")}
        />
        <Tile
          label={t("admin.revenue_paying_teams_label")}
          value={data.paying_teams.toLocaleString()}
        />
        <Tile
          label={t("admin.revenue_avg_payment")}
          value={
            data.paying_teams > 0
              ? formatMoney(Math.round(data.total_cents / data.paying_teams), data.currency)
              : "—"
          }
          hint={t("admin.revenue_avg_lead")}
        />
      </div>

      <div className="eop-grid grid grid-cols-12 gap-[14px]">
        <div className="eop-card col-span-12 min-[1181px]:col-span-5 min-w-0">
          <div className="card-head">
            <div>
              <div className="card-title">{t("admin.revenue_by_plan")}</div>
              <div className="card-sub">{t("admin.revenue_by_plan_sub")}</div>
            </div>
          </div>
          <div className="flex flex-col gap-2.5">
            {Object.entries(data.by_plan).map(([plan, count]) => {
              const total = Object.values(data.by_plan).reduce((acc, n) => acc + n, 0);
              const pct = total > 0 ? Math.round((count / total) * 100) : 0;
              return (
                <div
                  key={plan}
                  className="grid items-center gap-3 py-2"
                  style={{
                    gridTemplateColumns: "90px 1fr 80px",
                    borderBottom: "1px solid hsl(var(--border))",
                  }}
                >
                  <span className="font-mono text-[12px] uppercase tracking-widest2">{plan}</span>
                  <div
                    className="h-2 rounded-sm"
                    style={{ background: "rgba(255,255,255,0.05)", overflow: "hidden" }}
                  >
                    <div
                      style={{
                        width: `${pct}%`,
                        height: "100%",
                        background: planColor(plan),
                      }}
                    />
                  </div>
                  <span className="font-mono text-[12px] text-right">
                    {count} · {pct}%
                  </span>
                </div>
              );
            })}
          </div>
        </div>

        <div className="eop-card col-span-12 min-[1181px]:col-span-7 min-w-0">
          <div className="card-head">
            <div>
              <div className="card-title">{t("admin.revenue_recent_payments")}</div>
              <div className="card-sub">
                {t("admin.revenue_recent_sub", { n: data.recent.length })}
              </div>
            </div>
          </div>
          {data.recent.length === 0 ? (
            <div className="text-[13px] text-muted-foreground py-3">
              {t("admin.revenue_no_payments")}
            </div>
          ) : (
            <div className="log-table">
              <div className="log-row head">
                <span>{t("admin.payment_col_paid_at") || "PAID"}</span>
                <span>{t("admin.payment_col_team") || "TEAM"}</span>
                <span>{t("admin.payment_col_method") || "METHOD"}</span>
                <span style={{ textAlign: "right" }}>
                  {t("admin.payment_col_amount") || "AMOUNT"}
                </span>
                <span />
              </div>
              {data.recent.map((p) => (
                <PaymentRow key={p.id} p={p} tz={tz} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
function PaymentRow({ p, tz }: { p: RevenuePayment; tz: string }) {
  return (
    <div className="log-row">
      <span className="log-time">{formatDate(p.paid_at, tz)}</span>
      <span className="log-file truncate">{p.team_name}</span>
      <span className="log-lines">{p.method ?? "—"}</span>
      <span style={{ textAlign: "right" }} className="font-mono tabular-nums">
        {p.amount_cents !== undefined && p.currency ? formatMoney(p.amount_cents, p.currency) : "—"}
      </span>
      <span />
    </div>
  );
}
function Tile({ label, value, hint }: { label: string; value: string; hint?: string }) {
  return (
    <div className="kpi">
      <span className="kpi-label">{label}</span>
      <span className="kpi-value">{value}</span>
      {hint && <span className="kpi-delta flat">{hint}</span>}
    </div>
  );
}
function formatMoney(cents: number, currency: string): string {
  const major = (cents / 100).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  return `${major} ${currency.toUpperCase()}`;
}
function planColor(plan: string): string {
  switch (plan) {
    case "enterprise":
      return "#c084fc";
    case "business":
      return "hsl(var(--accent))";
    case "pro":
      return "#60a5fa";
    case "team":
      return "#60a5fa";
    case "free":
    default:
      return "#4ade80";
  }
}
