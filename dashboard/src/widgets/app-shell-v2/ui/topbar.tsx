// Topbar — breadcrumb + search + date-picker + icon-buttons + pause-btn.
// Mobile: collapses menu-btn, hides search.

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Bell, Menu, Pause, Search } from "lucide-react";
import { LangSwitch } from "../../../shared/ui/lang-switch";

const RANGES = ["24h", "7d", "30d", "90d"] as const;

export function Topbar({
  crumb,
  onMenuClick,
  onLogout,
}: {
  crumb: { section: string; now: string };
  onMenuClick?: () => void;
  onLogout: () => void;
}) {
  const { t } = useTranslation("common");
  const [range, setRange] = useState<(typeof RANGES)[number]>("7d");
  const [paused, setPaused] = useState(false);
  return (
    <header className="eop-topbar">
      {/* eslint-disable-next-line no-restricted-syntax -- icon-only toggle, IconButton from @eop/ui doesn't carry custom styling */}
      <button
        type="button"
        onClick={onMenuClick}
        className="md:hidden grid place-items-center w-9 h-9 rounded-lg border"
        style={{ borderColor: "hsl(var(--border))" }}
        aria-label={t("topbar.menu_toggle")}
      >
        <Menu className="h-5 w-5" />
      </button>

      <div className="crumb hidden sm:flex">
        <span>{crumb.section}</span>
        <span className="sep">/</span>
        <span className="now">{crumb.now}</span>
      </div>

      <div className="search-bar">
        <Search className="h-4 w-4" />
        <span>{t("topbar.search")}…</span>
        <span className="kbd">⌘ K</span>
      </div>

      <div className="ml-auto flex items-center gap-2.5">
        <div className="date-pick">
          {RANGES.map((r) => (
            // eslint-disable-next-line no-restricted-syntax -- date-picker tab control (micro-control inside topbar)
            <button
              key={r}
              type="button"
              className={range === r ? "on" : ""}
              onClick={() => setRange(r)}
            >
              {r}
            </button>
          ))}
        </div>

        <LangSwitch />

        {/* eslint-disable-next-line no-restricted-syntax -- icon-only notif button */}
        <button
          type="button"
          aria-label={t("topbar.notifications")}
          className="relative w-9 h-9 grid place-items-center rounded-lg border transition-colors hover:bg-foreground/5"
          style={{ borderColor: "hsl(var(--border))" }}
        >
          <Bell className="h-4 w-4" />
        </button>

        {/* eslint-disable-next-line no-restricted-syntax -- bespoke pause-btn styling */}
        <button
          type="button"
          className="pause-btn hidden md:inline-flex"
          onClick={() => setPaused((p) => !p)}
          title={paused ? t("topbar.resume_title") : t("topbar.pause_title")}
        >
          {paused ? null : <span className="dot" />}
          <Pause className="h-3.5 w-3.5" />
          {paused ? t("topbar.paused") : t("topbar.tracking")}
        </button>

        {/* eslint-disable-next-line no-restricted-syntax -- inline text-only link-style action */}
        <button
          type="button"
          onClick={onLogout}
          className="text-sm text-muted-foreground hover:text-foreground transition-colors px-2"
        >
          {t("actions.logout")}
        </button>
      </div>
    </header>
  );
}
