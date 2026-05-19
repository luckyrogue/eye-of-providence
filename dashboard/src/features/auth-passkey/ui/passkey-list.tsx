// Settings → Passkeys → list rows.
//
// Empty state matches the copy in product-copy/oauth-and-passkey.md. The
// revoke button warns destructive intent through the shared useConfirm
// dialog; the bottom toast confirms success / surfaces backend errors.
import { useTranslation } from "react-i18next";
import { Button, useConfirm } from "@eop/ui";
import { Fingerprint, Trash2 } from "lucide-react";
import { useMutationToast } from "@/shared/hooks";
import { usePasskeys, useRevokePasskey } from "../api";
import type { Passkey } from "@/entities/user";

function formatRelative(iso: string | null | undefined, locale: string): string | null {
  if (!iso) return null;
  return new Date(iso).toLocaleString(locale, {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function PasskeyList() {
  const { t, i18n } = useTranslation("common");
  const passkeys = usePasskeys();
  const revoke = useRevokePasskey();
  const confirm = useConfirm();
  const runToast = useMutationToast();
  const locale = i18n.resolvedLanguage ?? "ru";

  async function onRevoke(p: Passkey) {
    const isLast = (passkeys.data?.length ?? 0) <= 1;
    const ok = await confirm({
      title: t("settings.passkey_revoke"),
      description: isLast
        ? t("settings.passkey_revoke_last_lead")
        : t("settings.passkey_revoke_lead"),
      destructive: true,
      confirmText: t("settings.passkey_revoke"),
    });
    if (!ok) return;
    await runToast(revoke.mutateAsync(p.id), {});
  }

  if (passkeys.isLoading) {
    return <p className="text-sm text-muted-foreground">{t("actions.loading")}</p>;
  }

  const items = passkeys.data ?? [];
  if (items.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("settings.passkey_empty")}</p>;
  }

  return (
    <ul className="space-y-2">
      {items.map((p) => {
        const lastUsed = formatRelative(p.last_used_at, locale);
        const added = formatRelative(p.created_at, locale);
        return (
          <li
            key={p.id}
            className="flex items-start justify-between gap-3 rounded-md border px-3 py-2"
          >
            <div className="min-w-0 flex items-start gap-2">
              <Fingerprint className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" aria-hidden />
              <div className="min-w-0">
                <div className="font-medium text-sm truncate">{p.nickname}</div>
                <div className="text-xs text-muted-foreground space-x-2">
                  {lastUsed && <span>{t("settings.passkey_last_used", { date: lastUsed })}</span>}
                  {added && <span>· {t("settings.passkey_added", { date: added })}</span>}
                </div>
              </div>
            </div>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => onRevoke(p)}
              disabled={revoke.isPending}
              aria-label={t("settings.passkey_revoke")}
            >
              <Trash2 className="h-3.5 w-3.5" aria-hidden />
            </Button>
          </li>
        );
      })}
    </ul>
  );
}
