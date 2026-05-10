import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Input } from "@eop/ui";
import { ArrowRight, Building2, Loader2 } from "lucide-react";

export function CompanyStep({
  name,
  setName,
  busy,
  onSubmit,
}: {
  name: string;
  setName: (v: string) => void;
  busy: boolean;
  onSubmit: () => void;
}) {
  const { t } = useTranslation("onboarding");
  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Building2 className="h-5 w-5" />
          {t("company.title")}
        </CardTitle>
        <CardDescription>{t("company.lead")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Input
          autoFocus
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && !busy && name.trim() && onSubmit()}
          placeholder={t("company.placeholder")}
          maxLength={100}
          className="w-full text-base"
        />
        <div className="flex justify-end">
          <Button onClick={onSubmit} disabled={busy || !name.trim()}>
            {busy ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
            {t("company.submit")} <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
