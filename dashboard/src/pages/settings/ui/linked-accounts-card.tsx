// Settings → Linked accounts card.
//
// Lists existing identities (provider + email + connected date + unlink) and
// renders "Connect <provider>" buttons for the diff between
// authConfig.providers and what the user already linked.
//
// Connecting uses the same OAuth start URL as the login page but with a
// `return_to=/settings` hint baked in: the backend forwards this to its
// cookie handoff, so on success we land on `/auth/complete?return_to=/settings`
// and the new identity row shows up immediately.
import type { ComponentType } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  useConfirm,
} from "@eop/ui";
import { Link2, Trash2 } from "lucide-react";
import type { Identity, OAuthProvider } from "@/entities/user";
import { useAuthConfig } from "@/entities/user";
import { useMutationToast } from "@/shared/hooks";
import { AppleGlyph, GithubGlyph, GoogleGlyph } from "@/shared/ui/oauth-glyphs";
import { useIdentities, useUnlinkIdentity } from "../api";

const API_BASE = import.meta.env.VITE_BACKEND_URL || "https://eop-api.rysdavletov.org";

const GLYPHS: Record<OAuthProvider, ComponentType<{ className?: string }>> = {
  github: GithubGlyph,
  google: GoogleGlyph,
  apple: AppleGlyph,
};

function startLink(provider: OAuthProvider) {
  const returnTo = encodeURIComponent("/settings");
  window.location.href = `${API_BASE}/v1/auth/${provider}/login?return_to=${returnTo}`;
}

export function LinkedAccountsCard() {
  const { t, i18n } = useTranslation(["common", "auth"]);
  const authConfig = useAuthConfig();
  const identities = useIdentities();
  const unlink = useUnlinkIdentity();
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const locale = i18n.resolvedLanguage ?? "ru";

  const enabledProviders: OAuthProvider[] = authConfig.data?.providers ?? [];
  const linked: Identity[] = identities.data ?? [];
  const linkedProviders = new Set(linked.map((i) => i.provider));
  const availableProviders = enabledProviders.filter((p) => !linkedProviders.has(p));

  async function onUnlink(identity: Identity) {
    const ok = await confirm({
      title: t("common:settings.linked_unlink"),
      description: t("common:settings.linked_unlink_lead"),
      destructive: true,
      confirmText: t("common:settings.linked_unlink"),
    });
    if (!ok) return;
    try {
      await unlink.mutateAsync(identity.id);
      await runToast(Promise.resolve(true), {});
    } catch (e) {
      const err = e as { code?: string; message?: string };
      const msg =
        err.code === "last_auth_factor"
          ? t("common:settings.linked_lockout")
          : err.message || t("auth:passkey.error_generic");
      await runToast(Promise.reject(new Error(msg)), {});
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Link2 className="h-4 w-4 text-muted-foreground" />
          <CardTitle>{t("common:settings.linked_title")}</CardTitle>
        </div>
        <CardDescription>{t("common:settings.linked_lead")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {identities.isLoading ? (
          <p className="text-sm text-muted-foreground">{t("common:actions.loading")}</p>
        ) : linked.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("common:settings.linked_empty")}</p>
        ) : (
          <ul className="space-y-2">
            {linked.map((identity) => {
              const Glyph = GLYPHS[identity.provider];
              const connected = new Date(identity.created_at).toLocaleDateString(locale, {
                year: "numeric",
                month: "short",
                day: "2-digit",
              });
              return (
                <li
                  key={identity.id}
                  className="flex items-start justify-between gap-3 rounded-md border px-3 py-2"
                >
                  <div className="min-w-0 flex items-start gap-2">
                    {Glyph && <Glyph className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />}
                    <div className="min-w-0">
                      <div className="font-medium text-sm truncate">
                        {t(`auth:oauth.provider.${identity.provider}` as const)}
                      </div>
                      <div className="text-xs text-muted-foreground truncate">
                        {identity.email ?? identity.subject}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {t("common:settings.linked_connected", { date: connected })}
                      </div>
                    </div>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => onUnlink(identity)}
                    disabled={unlink.isPending}
                    aria-label={t("common:settings.linked_unlink")}
                  >
                    <Trash2 className="h-3.5 w-3.5" aria-hidden />
                  </Button>
                </li>
              );
            })}
          </ul>
        )}

        {availableProviders.length > 0 && (
          <div className="border-t pt-3">
            <div className="text-sm font-medium mb-2">{t("common:settings.linked_connect")}</div>
            <div className="flex flex-col gap-2 sm:flex-row">
              {availableProviders.map((provider) => {
                const Glyph = GLYPHS[provider];
                return (
                  <Button
                    key={provider}
                    type="button"
                    variant="outline"
                    onClick={() => startLink(provider)}
                    className="w-full"
                  >
                    <Glyph className="h-4 w-4" />
                    <span>
                      {t("auth:oauth.continue_with", {
                        provider: t(`auth:oauth.provider.${provider}` as const),
                      })}
                    </span>
                  </Button>
                );
              })}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
