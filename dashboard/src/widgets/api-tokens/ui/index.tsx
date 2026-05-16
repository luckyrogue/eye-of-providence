// API tokens management — Settings page section.
//
// Тонкий widget: список токенов из entities/user + создание/отзыв через
// features. Plaintext-secret виден ровно один раз — feature сама держит
// двухшаговую модалку.

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { KeyRound, Plus } from "lucide-react";
import { useTokens } from "../../../entities/user";
import { CreateTokenDialog } from "../../../features/api-token-create";
import { TokenRow } from "./token-row";

export function APITokensWidget() {
  const { t } = useTranslation("developer");
  const tokens = useTokens();
  const [showCreate, setShowCreate] = useState(false);

  return (
    <Card>
      <CardHeader className="flex-col sm:flex-row sm:items-start sm:justify-between gap-3 sm:gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <KeyRound className="h-4 w-4 text-muted-foreground shrink-0" />
            <CardTitle>{t("tokens_title")}</CardTitle>
          </div>
          <CardDescription className="mt-1">{t("tokens_lead")}</CardDescription>
        </div>
        <Button
          type="button"
          size="sm"
          variant="default"
          onClick={() => setShowCreate(true)}
          className="w-full shrink-0 justify-center gap-2 sm:h-10 sm:w-10 sm:min-w-10 sm:max-w-10 sm:gap-0 sm:px-0"
          title={t("tokens_create")}
          aria-label={t("tokens_create")}
        >
          <Plus className="h-4 w-4 shrink-0" aria-hidden />
          <span className="sm:hidden">{t("tokens_create")}</span>
        </Button>
      </CardHeader>
      <CardContent>
        {tokens.isPending ? (
          <p className="text-sm text-muted-foreground">…</p>
        ) : !tokens.data || tokens.data.length === 0 ? null : (
          <ul className="divide-y">
            {tokens.data.map((tok) => (
              <TokenRow key={tok.id} token={tok} />
            ))}
          </ul>
        )}
      </CardContent>

      <CreateTokenDialog open={showCreate} onClose={() => setShowCreate(false)} />
    </Card>
  );
}
