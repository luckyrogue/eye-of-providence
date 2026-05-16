import { Trans, useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { ArrowRight, Download, SkipForward } from "lucide-react";

export function InstallStep({
  onSkip,
  onContinue,
}: {
  onSkip: () => void;
  onContinue: () => void;
}) {
  const { t } = useTranslation("onboarding");
  const platforms = [
    { os: "macOS", file: ".dmg", note: t("install.macos_note"), href: "/downloads/eop-mac.dmg" },
    {
      os: "Windows",
      file: ".msi",
      note: t("install.windows_note"),
      href: "/downloads/eop-win.msi",
    },
    {
      os: "VS Code",
      file: "extension",
      note: t("install.vscode_note"),
      href: "https://marketplace.visualstudio.com",
    },
  ];
  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Download className="h-5 w-5" />
          {t("install.title")}
        </CardTitle>
        <CardDescription>{t("install.lead")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
          {platforms.map((p) => (
            <a
              key={p.os}
              href={p.href}
              target="_blank"
              rel="noreferrer"
              className="rounded-lg border bg-card p-4 text-center hover:bg-secondary card-hover"
            >
              <div className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">
                {p.os}
              </div>
              <div className="font-display font-bold mt-1">{p.file}</div>
              <div className="text-[11px] text-muted-foreground mt-1">{p.note}</div>
            </a>
          ))}
        </div>
        <div className="rounded-md border bg-muted/30 p-3 text-xs">
          <p className="text-muted-foreground">
            <Trans
              i18nKey="onboarding:install.extras"
              components={{
                a: <a href="/docs/install" className="underline hover:text-foreground" />,
              }}
            />
          </p>
        </div>
        <div className="flex justify-between pt-2">
          <Button variant="ghost" onClick={onSkip} className="text-muted-foreground">
            <SkipForward className="h-3.5 w-3.5 mr-1.5" />
            {t("install.skip")}
          </Button>
          <Button onClick={onContinue}>
            {t("install.done")} <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
