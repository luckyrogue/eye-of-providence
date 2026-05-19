import { useTranslation } from "react-i18next";
import type { APIToken } from "@/entities/user";
import { RevokeTokenButton } from "./revoke-token-button";

export function TokenRow({ token }: { token: APIToken }) {
  const { t } = useTranslation("developer");
  const lastUsed = token.last_used_at
    ? t("tokens_last_used", { at: new Date(token.last_used_at).toLocaleString() })
    : t("tokens_never_used");
  const expires = token.expires_at
    ? t("tokens_expires", { at: new Date(token.expires_at).toLocaleDateString() })
    : t("tokens_no_expiry");

  return (
    <li className="flex items-center justify-between py-3 gap-4">
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="font-medium text-sm">{token.name}</span>
          <code className="font-mono text-xs px-1.5 py-0.5 rounded bg-muted">{token.prefix}…</code>
          <span className="text-[11px] uppercase tracking-wider text-muted-foreground">
            {token.scope}
          </span>
        </div>
        <div className="text-xs text-muted-foreground mt-1">
          {lastUsed} · {expires}
        </div>
      </div>
      <RevokeTokenButton token={token} />
    </li>
  );
}
