// CMS-lite editor for a single slug × locale row.
//
// Layout:
//   - Header: slug + locale tabs + Save Draft / Publish / Revert / Preview.
//   - Monaco JSON editor with per-slug JSON Schema for live validation.
//   - Optional preview iframe (sandbox=""): renders /?preview=<slug>.
//
// Concurrency:
//   - PUT carries If-Match=<last-known etag>. On 412 we surface a localised
//     "conflict" toast but keep the unsaved JSON in memory so the admin
//     can hit Save again after refreshing.

import { lazy, Suspense, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, useConfirm } from "@eop/ui";
import { Loader2 } from "lucide-react";
import {
  CONTENT_EXAMPLES,
  CONTENT_LOCALES,
  CONTENT_SCHEMAS,
  type ContentLocale,
  type ContentSlug,
} from "../../../shared/content";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { ConcurrentEditError, useContentBlock, useRevertContent, useSaveContent } from "../api";

// Lazy-load Monaco — keeps the admin bundle lean. We use the same
// component as the email-template editor (Phase 3 pattern).
const MonacoEditor = lazy(() =>
  import("@monaco-editor/react").then((m) => ({ default: m.Editor })),
);

type EditorErrorState = "none" | "schema" | "syntax" | "conflict";

export function ContentBlockEditor({ slug }: { slug: ContentSlug }) {
  const { t } = useTranslation("app");
  const confirm = useConfirm();
  const runToast = useMutationToast();
  const [locale, setLocale] = useState<ContentLocale>("en");
  const detail = useContentBlock(slug, locale);
  const save = useSaveContent();
  const revert = useRevertContent();

  // Editor state — kept separately from react-query cache so unsaved edits
  // survive a background refetch.
  const [json, setJson] = useState<string>("");
  const [dirty, setDirty] = useState(false);
  const [editorError, setEditorError] = useState<EditorErrorState>("none");

  // Reset editor when slug/locale or backing detail changes.
  useEffect(() => {
    if (!detail.data) return;
    // Prefer draft over published when the row has both (an in-flight
    // draft is what the admin wants to keep editing).
    const seed = detail.data.draft_content ?? detail.data.content ?? CONTENT_EXAMPLES[slug];
    setJson(JSON.stringify(seed, null, 2));
    setDirty(false);
    setEditorError("none");
  }, [detail.data, slug, locale]);

  const schema = CONTENT_SCHEMAS[slug];

  // Configure Monaco's JSON validation with the per-slug schema. Uses a
  // synthetic URI so multiple slugs don't collide in the schemas list.
  const monacoUri = useMemo(() => `inmemory://eop-cms/${slug}-${locale}.json`, [slug, locale]);

  // Monaco 0.55 moved JSON language services to the top-level `json`
  // namespace (the legacy `languages.json` is deprecated). We type the
  // arg with a minimal shape so a future Monaco upgrade doesn't silently
  // break the schema wiring.
  type MonacoLike = {
    json: {
      jsonDefaults: {
        setDiagnosticsOptions: (opts: {
          validate?: boolean;
          allowComments?: boolean;
          schemas?: Array<{ uri: string; fileMatch?: string[]; schema?: unknown }>;
        }) => void;
      };
    };
  };
  const onMonacoBeforeMount = (monaco: MonacoLike) => {
    monaco.json.jsonDefaults.setDiagnosticsOptions({
      validate: true,
      allowComments: false,
      schemas: [
        {
          uri: `inmemory://eop-cms-schema/${slug}.json`,
          fileMatch: [monacoUri],
          schema: schema,
        },
      ],
    });
  };

  const parse = (): { ok: true; value: unknown } | { ok: false; error: string } => {
    try {
      return { ok: true, value: JSON.parse(json) };
    } catch (e) {
      return { ok: false, error: e instanceof Error ? e.message : String(e) };
    }
  };

  async function doSave(publish: boolean) {
    const parsed = parse();
    if (!parsed.ok) {
      setEditorError("syntax");
      return;
    }
    setEditorError("none");
    try {
      const r = await save.mutateAsync({
        slug,
        locale,
        payload: { content: parsed.value, publish },
        etag: detail.data?.etag ?? null,
      });
      // After save: the JSON we just typed is now the source of truth.
      // Reset dirty flag against the response so cosmetic re-saves are no-ops.
      setDirty(false);
      await runToast(Promise.resolve(r), {
        success: publish
          ? t("admin.content.publish_success", { defaultValue: "Published" })
          : t("admin.content.save_success", { defaultValue: "Draft saved" }),
      });
    } catch (e) {
      if (e instanceof ConcurrentEditError) {
        setEditorError("conflict");
        await runToast(Promise.reject(e), {
          error: t("admin.content.conflict_error", {
            defaultValue: "Content was modified by someone else",
          }),
        });
        return;
      }
      await runToast(Promise.reject(e), {
        error: t("admin.content.schema_error", {
          defaultValue: "Invalid format: see editor highlights",
        }),
      });
    }
  }

  async function doRevert() {
    const ok = await confirm({
      title: t("admin.content.confirm_revert_title", {
        defaultValue: "Revert to bundled default?",
      }),
      description: t("admin.content.confirm_revert_lead", {
        defaultValue:
          "Deletes the CMS row for this slug + locale. The landing page will render the bundled i18n string.",
      }),
      destructive: true,
      confirmText: t("admin.content.revert", { defaultValue: "Revert" }),
    });
    if (!ok) return;
    await runToast(revert.mutateAsync({ slug, locale }), {
      success: t("admin.content.save_success", { defaultValue: "Reverted" }),
    });
  }

  function openPreview() {
    const url = `/?preview=${encodeURIComponent(slug)}&locale=${encodeURIComponent(locale)}`;
    window.open(url, "_blank", "noopener,noreferrer");
  }

  return (
    <div className="grid grid-cols-1 xl:grid-cols-[1fr_minmax(0,420px)] gap-4">
      <div className="space-y-3 min-w-0">
        <header className="flex items-center justify-between gap-2 flex-wrap">
          <div className="min-w-0">
            <div className="text-base font-medium truncate">{slug}</div>
            <div className="text-xs text-muted-foreground font-mono flex items-center gap-2 mt-0.5">
              <span>{locale}</span>
              <BadgeForRow
                hasPublished={!!detail.data?.published_at}
                hasDraft={!!detail.data?.draft_content}
                isDefault={!detail.data?.content && !detail.data?.draft_content}
              />
            </div>
          </div>
          <div className="flex items-center gap-2 flex-wrap">
            <div
              className="inline-flex rounded-md border overflow-hidden text-[12px] font-mono"
              style={{ borderColor: "hsl(var(--border))" }}
            >
              {CONTENT_LOCALES.map((l) => (
                // eslint-disable-next-line no-restricted-syntax -- locale tabs
                <button
                  key={l}
                  type="button"
                  onClick={() => setLocale(l)}
                  className={`px-2.5 py-1 transition-colors ${
                    locale === l ? "bg-foreground/10" : "hover:bg-foreground/5"
                  }`}
                >
                  {l}
                </button>
              ))}
            </div>
            <Button type="button" variant="ghost" size="sm" onClick={openPreview}>
              {t("admin.content.preview", { defaultValue: "Preview" })}
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => void doRevert()}
              disabled={revert.isPending || (!detail.data?.content && !detail.data?.draft_content)}
            >
              {t("admin.content.revert", { defaultValue: "Revert" })}
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              disabled={save.isPending || !dirty}
              onClick={() => void doSave(false)}
            >
              {save.isPending && <Loader2 className="h-3.5 w-3.5 animate-spin mr-1.5" />}
              {t("admin.content.save_draft", { defaultValue: "Save draft" })}
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={save.isPending}
              onClick={() => void doSave(true)}
            >
              {save.isPending && <Loader2 className="h-3.5 w-3.5 animate-spin mr-1.5" />}
              {t("admin.content.publish", { defaultValue: "Publish" })}
            </Button>
          </div>
        </header>

        <Suspense
          fallback={<div className="h-[420px] rounded-md border bg-muted/30 animate-pulse" />}
        >
          <div
            className="rounded-md border overflow-hidden"
            style={{ borderColor: "hsl(var(--border))" }}
          >
            <MonacoEditor
              key={`${slug}-${locale}`}
              height="420px"
              language="json"
              defaultLanguage="json"
              path={monacoUri}
              value={json}
              theme="vs-dark"
              beforeMount={onMonacoBeforeMount}
              onChange={(v) => {
                setJson(v ?? "");
                setDirty(true);
              }}
              options={{
                minimap: { enabled: false },
                wordWrap: "on",
                fontSize: 13,
                lineNumbers: "on",
                scrollBeyondLastLine: false,
                tabSize: 2,
                formatOnPaste: true,
                formatOnType: true,
              }}
            />
          </div>
        </Suspense>

        {editorError === "syntax" && (
          <div className="text-xs" style={{ color: "hsl(var(--destructive))" }}>
            {t("admin.content.schema_error", {
              defaultValue: "Invalid JSON: see editor highlights",
            })}
          </div>
        )}
        {editorError === "conflict" && (
          <div className="text-xs" style={{ color: "hsl(var(--destructive))" }}>
            {t("admin.content.conflict_error", {
              defaultValue: "Content was modified by someone else",
            })}
          </div>
        )}
      </div>

      <aside
        className="rounded-md border overflow-hidden flex flex-col"
        style={{ borderColor: "hsl(var(--border))", minHeight: 420 }}
      >
        <div
          className="px-3 py-2 text-[11px] uppercase tracking-wide font-mono text-muted-foreground bg-muted/30 border-b"
          style={{ borderColor: "hsl(var(--border))" }}
        >
          {t("admin.content.preview_title", { defaultValue: "Live preview" })}
        </div>
        <iframe
          // sandbox="" — block scripts/cookies; just visual.
          title="cms-preview"
          sandbox=""
          className="flex-1 bg-white"
          src={`/?preview=${encodeURIComponent(slug)}&locale=${encodeURIComponent(locale)}`}
        />
      </aside>
    </div>
  );
}

function BadgeForRow({
  hasPublished,
  hasDraft,
  isDefault,
}: {
  hasPublished: boolean;
  hasDraft: boolean;
  isDefault: boolean;
}) {
  const { t } = useTranslation("app");
  if (isDefault) {
    return (
      <span
        className="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] uppercase tracking-wide"
        style={{ background: "hsl(var(--muted))" }}
      >
        {t("admin.content.default_badge", { defaultValue: "Default" })}
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1">
      {hasPublished && (
        <span
          className="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] uppercase tracking-wide"
          style={{ background: "hsl(var(--primary) / 0.15)", color: "hsl(var(--primary))" }}
        >
          {t("admin.content.published_badge", { defaultValue: "Published" })}
        </span>
      )}
      {hasDraft && (
        <span
          className="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] uppercase tracking-wide"
          style={{ background: "hsl(var(--accent) / 0.15)", color: "hsl(var(--accent))" }}
        >
          {t("admin.content.draft_badge", { defaultValue: "Draft" })}
        </span>
      )}
    </span>
  );
}
