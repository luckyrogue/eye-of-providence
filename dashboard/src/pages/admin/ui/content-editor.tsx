// CMS-lite editor page (Phase 4).
//
// Two-pane layout: slug tree on the left, single-slug editor on the right.
// The editor itself owns the locale tabs + Monaco + live preview iframe.

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { EmptyState } from "@eop/ui";
import { ContentBlockEditor, SlugTree } from "../../../features/admin-content";
import { CONTENT_SLUGS, type ContentSlug } from "../../../shared/content";

export function ContentEditorPage() {
  const { t } = useTranslation("app");
  const [selected, setSelected] = useState<ContentSlug | null>(CONTENT_SLUGS[0] ?? null);

  return (
    <div className="eop-card">
      <div className="card-head">
        <div>
          <div className="card-title">
            {t("admin.content.title", { defaultValue: "Content (CMS-lite)" })}
          </div>
          <div className="card-sub">
            {t("admin.content.intro", {
              defaultValue:
                "Edit landing-page copy without redeploying. Empty rows fall back to bundled i18n strings.",
            })}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-[280px_1fr] gap-4">
        <SlugTree selected={selected} onSelect={setSelected} />

        <section
          className="rounded-md border p-4 min-h-[520px]"
          style={{ borderColor: "hsl(var(--border))" }}
        >
          {selected ? (
            <ContentBlockEditor key={selected} slug={selected} />
          ) : (
            <EmptyState
              eyebrow={t("admin.content.empty_eyebrow", { defaultValue: "No slug" })}
              title={t("admin.content.empty_title", {
                defaultValue: "Select a slug on the left",
              })}
            />
          )}
        </section>
      </div>
    </div>
  );
}
