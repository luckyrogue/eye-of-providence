import { useTranslation } from "react-i18next";
import { TYPE_ICON } from "../model/icon-map";
import type { ChangelogEntry } from "../model/types";

export function Row({ entry }: { entry: ChangelogEntry }) {
  const { t } = useTranslation("changelog");
  const cfg = TYPE_ICON[entry.type];
  const Icon = cfg.icon;
  return (
    <li className="flex items-start gap-3 rounded-md border bg-muted/20 p-3 transition-colors hover:bg-muted/40">
      <Icon className={`h-4 w-4 mt-0.5 shrink-0 ${cfg.color}`} />
      <div className="min-w-0 flex-1">
        <div className="flex items-baseline gap-2 flex-wrap">
          <span className="text-[11px] font-mono uppercase tracking-wider text-muted-foreground">
            {t(`type.${entry.type}` as const)}
          </span>
          {entry.scope && (
            <span className="text-[11px] font-mono px-1.5 py-0.5 rounded bg-muted">
              {entry.scope}
            </span>
          )}
          {entry.breaking && (
            <span className="text-[10px] uppercase font-semibold tracking-wider px-1.5 py-0.5 rounded bg-red-500/10 text-red-600 dark:text-red-400">
              {t("breaking")}
            </span>
          )}
        </div>
        <p className="text-sm mt-0.5">{entry.summary}</p>
      </div>
      <code className="text-[10px] font-mono text-muted-foreground tabular-nums shrink-0">
        {entry.hash}
      </code>
    </li>
  );
}
