import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { ArrowRight, Download, SkipForward } from "lucide-react";

const PLATFORMS = [
  { os: "macOS", file: ".dmg", note: "10.15+", href: "/downloads/eop-mac.dmg" },
  { os: "Windows", file: ".msi", note: "10/11", href: "/downloads/eop-win.msi" },
  { os: "VS Code", file: "extension", note: "marketplace", href: "https://marketplace.visualstudio.com" },
];

export function InstallStep({ onSkip, onContinue }: { onSkip: () => void; onContinue: () => void }) {
  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Download className="h-5 w-5" />
          Установите агент
        </CardTitle>
        <CardDescription>
          Без агента события не будут собираться. Можно установить позже — установка займёт ~2 минуты.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
          {PLATFORMS.map((p) => (
            <a
              key={p.os}
              href={p.href}
              target="_blank"
              rel="noreferrer"
              className="rounded-lg border bg-card p-4 text-center hover:bg-secondary card-hover"
            >
              <div className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">{p.os}</div>
              <div className="font-display font-bold mt-1">{p.file}</div>
              <div className="text-[11px] text-muted-foreground mt-1">{p.note}</div>
            </a>
          ))}
        </div>
        <div className="rounded-md border bg-muted/30 p-3 text-xs">
          <p className="text-muted-foreground">
            Также есть browser extension и git-hooks — детали в{" "}
            <a href="/docs/install" className="underline hover:text-foreground">документации</a>.
          </p>
        </div>
        <div className="flex justify-between pt-2">
          <Button variant="ghost" onClick={onSkip} className="text-muted-foreground">
            <SkipForward className="h-3.5 w-3.5 mr-1.5" />
            Установлю позже
          </Button>
          <Button onClick={onContinue}>
            Готово <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
