import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { X } from "lucide-react";
import { Button, IconButton, Tabs, TabsList, TabsTrigger } from "@eop/ui";
import type { AdminTeam } from "../../../entities/admin";
import { TeamFlagsForm } from "../../../features/admin-team-flags";
import { PlanLimitsForm } from "../../../features/admin-plan-overrides";
import { formatDate } from "../../../shared/lib/tz";
type DrawerTab = "info" | "flags" | "limits";
const TABS: {
  id: DrawerTab;
  labelKey: string;
}[] = [
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
  useEffect(() => {
    if (team) setTab("info");
  }, [team]);
  if (!open || !team) return null;
  return (
    <div className="fixed inset-0 z-50" role="dialog" aria-modal="true" aria-label={team.name}>
      <Button
        type="button"
        variant="ghost"
        aria-label={t("admin.team_drawer.close")}
        onClick={onClose}
        className="absolute inset-0 h-auto min-h-0 w-full rounded-none bg-background/70 p-0 hover:bg-background/70 backdrop-blur-sm animate-fade-in"
      />

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
          <IconButton title={t("admin.team_drawer.close")} onClick={onClose}>
            <X className="h-4 w-4" />
          </IconButton>
        </header>

        <Tabs
          value={tab}
          onValueChange={(v) => setTab(v as DrawerTab)}
          className="flex min-h-0 flex-1 flex-col"
        >
          <TabsList
            className="h-auto w-full shrink-0 justify-start gap-0 rounded-none border-b bg-transparent p-0 px-4 pt-3"
            style={{ borderColor: "hsl(var(--border))" }}
          >
            {TABS.map((it) => (
              <TabsTrigger
                key={it.id}
                value={it.id}
                className="rounded-none border-0 border-b-2 border-transparent px-3 py-2 text-sm shadow-none ring-offset-0 focus-visible:ring-0 data-[state=active]:border-foreground data-[state=active]:bg-transparent data-[state=active]:text-foreground data-[state=inactive]:text-muted-foreground data-[state=inactive]:hover:bg-transparent data-[state=inactive]:hover:text-foreground"
              >
                {t(it.labelKey)}
              </TabsTrigger>
            ))}
          </TabsList>

          <div className="flex-1 overflow-y-auto p-4">
            {tab === "info" && (
              <div className="space-y-3">
                <Row label={t("admin.team_drawer.info_plan")} value={team.subscription_plan} />
                <Row
                  label={t("admin.team_drawer.info_subscription_until")}
                  value={team.subscription_until ? formatDate(team.subscription_until, tz) : "—"}
                />
                <Row
                  label={t("admin.team_drawer.info_members")}
                  value={String(team.member_count)}
                />
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
        </Tabs>
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
