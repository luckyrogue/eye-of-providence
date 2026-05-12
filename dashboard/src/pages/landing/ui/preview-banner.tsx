// Non-dismissable banner shown when `?preview=<slug>` is in the URL.
// Only renders for super_admins (silent for anonymous visitors — backend
// rejects the admin endpoint and the slug falls back to its published
// row, so the banner would be misleading).
//
// Exit button strips the `preview` query param via react-router.

import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";
import { usePreviewContext } from "../../../shared/content";

export function PreviewBanner() {
  const { t } = useTranslation("app");
  const { previewSlug } = usePreviewContext();
  const [params, setParams] = useSearchParams();

  if (!previewSlug) return null;

  const exit = () => {
    const next = new URLSearchParams(params);
    next.delete("preview");
    setParams(next, { replace: true });
  };

  return (
    <div
      role="status"
      aria-live="polite"
      className="sticky top-0 z-50 w-full px-4 py-2 text-[13px] font-mono flex items-center justify-between gap-3"
      style={{
        background: "hsl(var(--accent))",
        color: "hsl(var(--accent-foreground))",
      }}
    >
      <span className="truncate">
        {t("admin.content.preview_banner", {
          slug: previewSlug,
          defaultValue: `Preview: showing unpublished draft of \`${previewSlug}\`. Other content is live.`,
        })}
      </span>
      {/* eslint-disable-next-line no-restricted-syntax -- explicit exit button */}
      <button
        type="button"
        onClick={exit}
        className="shrink-0 inline-flex items-center rounded-md px-2.5 py-1 text-[12px] font-medium border"
        style={{
          borderColor: "hsl(var(--accent-foreground) / 0.4)",
          background: "transparent",
        }}
      >
        {t("admin.content.preview_exit", { defaultValue: "Exit preview" })}
      </button>
    </div>
  );
}
