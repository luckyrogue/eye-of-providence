// Email templates editor page. Two-column layout:
//   • left  — list of [key, locale] pairs (matrix view).
//   • right — Monaco editor + iframe preview.
//
// Outer state: currently selected (key, locale). Editor компонент
// сам подтягивает detail и кладёт назад через mutate.

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { EmptyState } from "@eop/ui";
import {
  useEmailTemplates,
  type EmailTemplateKey,
  type EmailTemplateLocale,
} from "@/entities/admin";
import { EmailTemplateEditor } from "./email-template-editor";

const ALL_KEYS: EmailTemplateKey[] = ["password_reset", "team_invite", "subscription_activated"];
const ALL_LOCALES: EmailTemplateLocale[] = ["en", "ru", "kk", "es"];

export function EmailTemplatesPage() {
  const { t } = useTranslation("app");
  const matrix = useEmailTemplates();
  const [selected, setSelected] = useState<{
    key: EmailTemplateKey;
    locale: EmailTemplateLocale;
  } | null>({ key: "password_reset", locale: "en" });

  // Маппинг [key, locale] → has_override для значка "overridden / default".
  const overrideMap = new Map<string, boolean>();
  for (const row of matrix.data ?? []) {
    overrideMap.set(`${row.key}:${row.locale}`, row.has_override);
  }

  return (
    <div className="eop-card">
      <div className="card-head">
        <div>
          <div className="card-title">{t("admin.email_templates.title")}</div>
          <div className="card-sub">{t("admin.email_templates.intro")}</div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-[260px_1fr] gap-4">
        <aside
          className="rounded-md border overflow-hidden"
          style={{ borderColor: "hsl(var(--border))" }}
        >
          {ALL_KEYS.map((key) => (
            <div key={key} className="border-b" style={{ borderColor: "hsl(var(--border))" }}>
              <div className="px-3 py-2 text-[11px] uppercase tracking-wide font-mono text-muted-foreground bg-muted/30">
                {t(`admin.email_templates.template.${key}`)}
              </div>
              <ul>
                {ALL_LOCALES.map((locale) => {
                  const active = selected?.key === key && selected.locale === locale;
                  const overridden = overrideMap.get(`${key}:${locale}`) ?? false;
                  return (
                    <li key={locale}>
                      {/* eslint-disable-next-line no-restricted-syntax */}
                      <button
                        type="button"
                        onClick={() => setSelected({ key, locale })}
                        className={`w-full flex items-center justify-between px-3 py-2 text-sm transition-colors hover:bg-foreground/5 ${
                          active ? "bg-foreground/10" : ""
                        }`}
                      >
                        <span className="font-mono">{locale}</span>
                        <span
                          className="text-[10px] uppercase tracking-wide rounded-full px-2 py-0.5"
                          style={{
                            background: overridden
                              ? "hsl(var(--primary) / 0.15)"
                              : "hsl(var(--muted))",
                            color: overridden
                              ? "hsl(var(--primary))"
                              : "hsl(var(--muted-foreground))",
                          }}
                        >
                          {overridden
                            ? t("admin.email_templates.badge.overridden")
                            : t("admin.email_templates.badge.default")}
                        </span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </aside>

        <section
          className="rounded-md border p-4 min-h-[480px]"
          style={{ borderColor: "hsl(var(--border))" }}
        >
          {!selected ? (
            <EmptyState
              eyebrow={t("admin.email_templates.empty_eyebrow")}
              title={t("admin.email_templates.empty_title")}
            />
          ) : (
            <EmailTemplateEditor
              key={`${selected.key}:${selected.locale}`}
              templateKey={selected.key}
              locale={selected.locale}
            />
          )}
        </section>
      </div>
    </div>
  );
}
