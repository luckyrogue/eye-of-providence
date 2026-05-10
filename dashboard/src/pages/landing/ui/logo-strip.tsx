import { useTranslation } from "react-i18next";
import { Activity, Brain, Code2, FileText, GitBranch, Sparkles } from "lucide-react";

const SOURCES = [
  { name: "VS Code", Icon: Code2 },
  { name: "Cursor", Icon: Sparkles },
  { name: "Claude Code", Icon: Brain },
  { name: "GitHub Copilot", Icon: GitBranch },
  { name: "Browser", Icon: Activity },
  { name: "CLI", Icon: FileText },
];

export function LogoStrip() {
  const { t } = useTranslation("landing");
  return (
    <section className="border-y bg-muted/20">
      <div className="mx-auto max-w-6xl px-6 py-10">
        <p className="text-center text-xs font-mono uppercase tracking-widest3 text-muted-foreground mb-6">
          {t("logo_strip.tagline")}
        </p>
        <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-4 text-sm text-muted-foreground">
          {SOURCES.map(({ name, Icon }) => (
            <span key={name} className="flex items-center gap-2 hover:text-foreground transition-colors">
              <Icon className="h-4 w-4" />
              <span className="font-medium">{name}</span>
            </span>
          ))}
        </div>
      </div>
    </section>
  );
}
