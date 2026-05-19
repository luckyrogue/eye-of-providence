// OAuth button row: renders enabled provider buttons side-by-side.
// Click → full redirect to `${VITE_BACKEND_URL}/v1/auth/${provider}/login`.
// SPA redirect is intentional — backend owns the OAuth round-trip.
import type { ComponentType } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import type { OAuthProvider } from "@/entities/user";
import { AppleGlyph, GithubGlyph, GoogleGlyph } from "@/shared/ui/oauth-glyphs";

const API_BASE = import.meta.env.VITE_BACKEND_URL || "https://eop-api.rysdavletov.org";

const GLYPHS: Record<OAuthProvider, ComponentType<{ className?: string }>> = {
  github: GithubGlyph,
  google: GoogleGlyph,
  apple: AppleGlyph,
};

function startOAuth(provider: OAuthProvider) {
  window.location.href = `${API_BASE}/v1/auth/${provider}/login`;
}

export function OAuthButtonRow({ providers }: { providers: OAuthProvider[] }) {
  const { t } = useTranslation("auth");
  if (!providers.length) return null;
  return (
    <div className="flex flex-col gap-2 sm:flex-row">
      {providers.map((provider) => {
        const Glyph = GLYPHS[provider];
        return (
          <Button
            key={provider}
            type="button"
            variant="outline"
            onClick={() => startOAuth(provider)}
            className="w-full"
          >
            <Glyph className="h-4 w-4" />
            <span>
              {t("oauth.continue_with", { provider: t(`oauth.provider.${provider}` as const) })}
            </span>
          </Button>
        );
      })}
    </div>
  );
}
