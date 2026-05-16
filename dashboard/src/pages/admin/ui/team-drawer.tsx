// Team admin drawer: right-slide panel. 3 tabs: Info | Flags | Limits.
//
// Don't pull @radix-ui directly (transitive dep through @eop/ui).
// Simple state-driven drawer with backdrop + Esc/click handlers and
// inert background scroll-lock.

import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { X } from "lucide-react";
import type { AdminTeam } from "../../../entities/admin";
import { TeamFlagsForm } from "../../../features/admin-team-flags";
import { PlanLimitsForm } from "../../../features/admin-plan-overrides";
import { formatDate } from "../../../shared/lib/tz";

type DrawerTab = "info" | "flags" | "limits";

const TABS: { id: DrawerTab; labelKey: string }[] = [
  { id: "info", labelKey: "admin.team_drawer.tab_info" },
  { id: "flags", labelKey: "admin.team_drawer.tab_flags" },
  { id: "limits", labelKey: "admin.team_drawer.tab_limits" },
];

export function TeamDrawer({
  team,
  tz,
  onClose,
}: {
  team: AdminTeam | null;
  tz: string;
  onClose: () => void;
}) {
  const { t } = useTranslation("app");
  const [tab, setTab] = useState<DrawerTab>("info");
  const open = !!team;

  // Esc to close + lock body scroll while open. Cleanup on unmount.
  useEffect(() => {
    if (!open) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      window.removeEventListener("keydown", onKey);
      document.body.style.overflow = prev;
    };
  }, [open, onClose]);

  // Reset tab to "info" каждый раз, когда открывается новая команда.
  useEffect(() => {
    if (team) setTab("info");
  }, [team?.id]); // eslint-disable-line react-hooks/exhaustive-deps

  if (!open || !team) return null;

  return (
    <div className="fixed inset-0 z-50" role="dialog" aria-modal="true" aria-label={team.name}>
      {/* Backdrop */}
      {/* eslint-disable-next-line no-restricted-syntax */}
      <button
        type="button"
        aria-label={t("admin.team_drawer.close")}
        onClick={onClose}
        className="absolute inset-0 bg-background/70 backdrop-blur-sm animate-fade-in"
      />

      {/* Panel */}
      <aside
        className="absolute right-0 top-0 h-full w-full max-w-[480px] border-l bg-card shadow-2xl flex flex-col animate-fade-in"
        style={{ borderColor: "hsl(var(--border))" }}
      >
        <header
          className="flex items-start justify-between gap-3 p-4 border-b"
          style={{ borderColor: "hsl(var(--border))" }}
        >
          <div className="min-w-0">
            <div className="font-display text-lg leading-none tracking-tight">{team.name}</div>
            <div className="text-xs text-muted-foreground mt-1 font-mono">{team.id}</div>
          </div>
          {/* eslint-disable-next-line no-restricted-syntax */}
          <button
            type="button"
            onClick={onClose}
            aria-label={t("admin.team_drawer.close")}
            className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        </header>

        <div
          className="flex gap-1 px-4 pt-3 border-b"
          style={{ borderColor: "hsl(var(--border))" }}
        >
          {TABS.map((it) => (
            // eslint-disable-next-line no-restricted-syntax
            <button
              key={it.id}
              type="button"
              role="tab"
              aria-selected={tab === it.id}
              onClick={() => setTab(it.id)}
              className={`px-3 py-2 text-sm border-b-2 -mb-px transition-colors ${
                tab === it.id
                  ? "border-foreground text-foreground"
                  : "border-transparent text-muted-foreground hover:text-foreground"
              }`}
            >
              {t(it.labelKey)}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto p-4">
          {tab === "info" && (
            <div className="space-y-3">
              <Row label={t("admin.team_drawer.info_plan")} value={team.subscription_plan} />
              <Row
                label={t("admin.team_drawer.info_subscription_until")}
                value={team.subscription_until ? formatDate(team.subscription_until, tz) : "—"}
              />
              <Row label={t("admin.team_drawer.info_members")} value={String(team.member_count)} />
              <Row label={t("admin.team_drawer.info_owner")} value={team.owner_email ?? "—"} />
              <Row
                label={t("admin.team_drawer.info_created")}
                value={formatDate(team.created_at, tz)}
              />
              {team.subscription_note && (
                <Row label={t("admin.team_drawer.info_note")} value={team.subscription_note} />
              )}
            </div>
          )}
          {tab === "flags" && <TeamFlagsForm teamID={team.id} />}
          {tab === "limits" && <PlanLimitsForm teamID={team.id} />}
        </div>
      </aside>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div
      className="flex items-start justify-between gap-2 border-b pb-2"
      style={{ borderColor: "hsl(var(--border))" }}
    >
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="text-sm font-mono break-all text-right">{value}</span>
    </div>
  );
}
