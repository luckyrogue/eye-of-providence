// Editor для одного email-шаблона (key + locale).
// RHF + zod. Monaco lazy-loaded. HTML preview через iframe sandbox.
// Required-variable hint и revert-to-default flow.

import { lazy, Suspense, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button, Input, useConfirm } from "@eop/ui";
import { Info, Loader2 } from "lucide-react";
import {
  useEmailTemplate,
  useRevertEmailTemplate,
  useSaveEmailTemplate,
  type EmailTemplateKey,
  type EmailTemplateLocale,
} from "../api";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

// Lazy-load Monaco — heavy chunk. Suspense fallback показывает skeleton.
const MonacoEditor = lazy(() =>
  import("@monaco-editor/react").then((m) => ({ default: m.Editor })),
);

// Per-template required variables — должен присутствовать в body_html.
const REQUIRED_VARS: Record<EmailTemplateKey, string[]> = {
  password_reset: ["reset_url"],
  team_invite: ["accept_url"],
  subscription_activated: ["plan_name"],
};

// Sample значения для preview (рендерятся в iframe sandbox).
const SAMPLE_VARS: Record<EmailTemplateKey, Record<string, string>> = {
  password_reset: {
    name: "Ada",
    reset_url: "https://eop.example/reset/sample",
    ip: "203.0.113.42",
    user_agent: "Mozilla/5.0",
  },
  team_invite: {
    inviter_name: "Ada",
    team_name: "Acme",
    accept_url: "https://eop.example/invite/sample",
    role: "member",
    expires_at: "May 19, 2026",
  },
  subscription_activated: {
    name: "Ada",
    plan_name: "Pro",
    next_invoice_date: "Jun 12, 2026",
    billing_url: "https://eop.example/billing",
  },
};

const schema = z.object({
  subject: z.string().min(1).max(998),
  body_html: z.string().min(1).max(262144),
  body_text: z.string().max(262144),
});

type FormValues = z.infer<typeof schema>;

export function EmailTemplateEditor({
  templateKey,
  locale,
}: {
  templateKey: EmailTemplateKey;
  locale: EmailTemplateLocale;
}) {
  const { t } = useTranslation("app");
  const confirm = useConfirm();
  const detail = useEmailTemplate(templateKey, locale);
  const save = useSaveEmailTemplate();
  const revert = useRevertEmailTemplate();
  const runToast = useMutationToast();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { subject: "", body_html: "", body_text: "" },
  });

  const { handleSubmit, watch, setValue, reset, formState } = form;
  const subject = watch("subject");
  const bodyHtml = watch("body_html");
  const bodyText = watch("body_text");

  useEffect(() => {
    if (detail.data) {
      reset({
        subject: detail.data.subject,
        body_html: detail.data.body_html,
        body_text: detail.data.body_text,
      });
    }
  }, [detail.data, reset]);

  const requiredVars = REQUIRED_VARS[templateKey];
  const missingVars = useMemo(
    () => requiredVars.filter((v) => !bodyHtml.includes(`{{${v}}}`)),
    [bodyHtml, requiredVars],
  );

  const previewHtml = useMemo(
    () => substituteVars(bodyHtml, SAMPLE_VARS[templateKey]),
    [bodyHtml, templateKey],
  );

  const [showPreview, setShowPreview] = useState(true);

  async function onSubmit(values: FormValues) {
    if (missingVars.length > 0) return;
    await runToast(save.mutateAsync({ key: templateKey, locale, payload: values }), {
      success: t("admin.email_templates.toast.saved"),
      error: t("admin.email_templates.toast.save_failed"),
    });
  }

  async function onRevert() {
    const ok = await confirm({
      title: t("admin.email_templates.cta.revert_confirm_title"),
      description: t("admin.email_templates.cta.revert_confirm_body"),
      destructive: true,
      confirmText: t("admin.email_templates.cta.revert"),
    });
    if (!ok) return;
    await runToast(revert.mutateAsync({ key: templateKey, locale }), {
      success: t("admin.email_templates.toast.reverted"),
      error: t("admin.email_templates.toast.revert_failed"),
    });
  }

  if (detail.isPending) {
    return (
      <div className="text-sm text-muted-foreground py-8 flex items-center gap-2">
        <Loader2 className="h-4 w-4 animate-spin" />
        {t("admin.loading")}
      </div>
    );
  }
  if (detail.isError) {
    return (
      <div className="text-sm text-destructive py-8">
        {detail.error?.message ?? t("admin.error_lead")}
      </div>
    );
  }
  if (!detail.data) return null;

  const isDefault = detail.data.is_default;

  return (
    <form className="space-y-4" onSubmit={handleSubmit(onSubmit)} noValidate>
      <header className="flex items-start justify-between gap-3 flex-wrap">
        <div className="min-w-0">
          <div className="text-base font-medium">
            {t(`admin.email_templates.template.${templateKey}`)}
          </div>
          <div className="text-xs text-muted-foreground font-mono">
            {templateKey} · {locale}{" "}
            <span
              className="ml-2 inline-flex items-center rounded-full px-2 py-0.5 text-[10px] uppercase tracking-wide"
              style={{ background: "hsl(var(--muted))" }}
            >
              {isDefault
                ? t("admin.email_templates.badge.default")
                : t("admin.email_templates.badge.overridden")}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button type="button" variant="ghost" size="sm" onClick={() => setShowPreview((s) => !s)}>
            {showPreview
              ? t("admin.email_templates.cta.hide_preview")
              : t("admin.email_templates.cta.preview")}
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => void onRevert()}
            disabled={isDefault || revert.isPending}
            title={t("admin.email_templates.cta.revert")}
          >
            {t("admin.email_templates.cta.revert")}
          </Button>
          <Button
            type="submit"
            size="sm"
            disabled={save.isPending || missingVars.length > 0 || !formState.isDirty}
          >
            {save.isPending && <Loader2 className="h-3.5 w-3.5 animate-spin mr-1.5" />}
            {t("admin.email_templates.cta.save")}
          </Button>
        </div>
      </header>

      <div>
        <label className="text-xs text-muted-foreground" htmlFor={`subject-${templateKey}`}>
          {t("admin.email_templates.field.subject")}
        </label>
        <Input
          id={`subject-${templateKey}`}
          value={subject}
          onChange={(e) => setValue("subject", e.target.value, { shouldDirty: true })}
          className="mt-1"
        />
      </div>

      <div>
        <div className="flex items-center justify-between mb-1">
          <label className="text-xs text-muted-foreground">
            {t("admin.email_templates.field.body_html")}
          </label>
          <span
            className="inline-flex items-center gap-1 text-[11px] text-muted-foreground"
            title={t(`admin.email_templates.vars.${templateKey}`)}
          >
            <Info className="h-3 w-3" />
            {t("admin.email_templates.vars.required_label")}
          </span>
        </div>
        <p className="text-[11px] text-muted-foreground mb-2 font-mono">
          {t(`admin.email_templates.vars.${templateKey}`)}
        </p>
        <Suspense fallback={<div className="h-72 rounded-md border bg-muted/30 animate-pulse" />}>
          <div
            className="rounded-md border overflow-hidden"
            style={{ borderColor: "hsl(var(--border))" }}
          >
            <MonacoEditor
              height="320px"
              defaultLanguage="html"
              language="html"
              value={bodyHtml}
              onChange={(v) => setValue("body_html", v ?? "", { shouldDirty: true })}
              theme="vs-dark"
              options={{
                minimap: { enabled: false },
                wordWrap: "on",
                fontSize: 13,
                lineNumbers: "on",
                scrollBeyondLastLine: false,
              }}
            />
          </div>
        </Suspense>
        {missingVars.length > 0 && (
          <div className="mt-2 text-xs" style={{ color: "hsl(var(--destructive))" }}>
            {missingVars.map((v) => (
              <div key={v}>
                {t("admin.email_templates.error.missing_variable", { variable: `{{${v}}}` })}
              </div>
            ))}
          </div>
        )}
      </div>

      <div>
        <label className="text-xs text-muted-foreground">
          {t("admin.email_templates.field.body_text")}
        </label>
        <Suspense
          fallback={<div className="h-40 rounded-md border bg-muted/30 animate-pulse mt-1" />}
        >
          <div
            className="rounded-md border overflow-hidden mt-1"
            style={{ borderColor: "hsl(var(--border))" }}
          >
            <MonacoEditor
              height="180px"
              defaultLanguage="plaintext"
              language="plaintext"
              value={bodyText}
              onChange={(v) => setValue("body_text", v ?? "", { shouldDirty: true })}
              theme="vs-dark"
              options={{
                minimap: { enabled: false },
                wordWrap: "on",
                fontSize: 13,
                lineNumbers: "off",
                scrollBeyondLastLine: false,
              }}
            />
          </div>
        </Suspense>
      </div>

      {showPreview && (
        <div>
          <div className="text-xs text-muted-foreground mb-1 flex items-center justify-between">
            <span>{t("admin.email_templates.preview.title")}</span>
            <span className="text-[10px]">{t("admin.email_templates.preview.sample_values")}</span>
          </div>
          <iframe
            title="email-preview"
            sandbox=""
            className="w-full h-72 rounded-md border bg-white"
            style={{ borderColor: "hsl(var(--border))" }}
            srcDoc={previewHtml}
          />
        </div>
      )}
    </form>
  );
}

// substituteVars — простая замена {{var}} → sample value. Используется
// только для preview; backend делает реальную подстановку с escaping.
function substituteVars(html: string, vars: Record<string, string>): string {
  return html.replace(/\{\{\s*(\w+)\s*\}\}/g, (_m, name: string) =>
    Object.prototype.hasOwnProperty.call(vars, name) ? vars[name] : `{{${name}}}`,
  );
}
