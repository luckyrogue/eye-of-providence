import { type ReactElement } from "react";
import { useTranslation } from "react-i18next";

type Item = {
  name: string;
  icon: string;
  ready?: boolean;
  soon?: boolean;
};

const ITEMS: Item[] = [
  { name: "VS Code", icon: "VSCode", ready: true },
  { name: "Cursor", icon: "Cursor", ready: true },
  { name: "Claude Code", icon: "Anthropic", ready: true },
  { name: "Copilot", icon: "Copilot", ready: true },
  { name: "OpenAI", icon: "OpenAI", ready: true },
  { name: "Gemini", icon: "Gemini", ready: true },
  { name: "Aider", icon: "Terminal", ready: true },
  { name: "Browser ext.", icon: "Browser", ready: true },
  { name: "GitHub", icon: "GitHub", ready: true },
  { name: "JetBrains", icon: "JetBrains", soon: true },
  { name: "macOS", icon: "Apple", ready: true },
  { name: "Windows", icon: "Windows", ready: true },
];

const Icons: Record<string, ReactElement> = {
  VSCode: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M17 3 7 11 3 8v8l4-3 10 8 4-2V5l-4-2Z" />
      <path d="M17 3 7 13" />
    </svg>
  ),
  Cursor: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M4 4l16 8-7 1-1 7L4 4Z" />
    </svg>
  ),
  Anthropic: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M14.5 4h3.7L24 20h-3.7l-1.3-3.4h-6.8L10.9 20H7.2L13 4h1.5zm-.7 3.7-2.4 6.4h4.8l-2.4-6.4zM3.7 4h3.7L13 20H9.3l-5.6-16z" />
    </svg>
  ),
  Copilot: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M4 12a8 8 0 0 1 16 0v3a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4v-3Z" />
      <circle cx="9" cy="13" r="1.5" fill="currentColor" />
      <circle cx="15" cy="13" r="1.5" fill="currentColor" />
    </svg>
  ),
  OpenAI: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M22.3 10c-.3-.8-.7-1.5-1.3-2 .3-.8.4-1.7.2-2.5-.4-1.8-1.9-3-3.7-3.2-.7-.1-1.5 0-2.2.3-.7-.6-1.6-1-2.6-1-2 0-3.7 1.4-4 3.4-.8.1-1.5.4-2.1.9-1.5 1.2-1.9 3.4-1 5 .3.7.7 1.3 1.3 1.7-.4 1.7.2 3.6 1.6 4.6 1 .8 2.3 1 3.6.7.7.6 1.6.9 2.6.9 2 0 3.7-1.4 4-3.4.7-.2 1.4-.5 2-1 1.5-1.3 1.8-3.4 1-5l.5-1.4z" />
    </svg>
  ),
  Gemini: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 2c0 5.5-4.5 10-10 10 5.5 0 10 4.5 10 10 0-5.5 4.5-10 10-10-5.5 0-10-4.5-10-10z" />
    </svg>
  ),
  Terminal: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="m6 9 3 3-3 3M12 16h6" />
    </svg>
  ),
  Browser: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <circle cx="12" cy="12" r="9" />
      <path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18" />
    </svg>
  ),
  GitHub: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 2a10 10 0 0 0-3.2 19.5c.5.1.7-.2.7-.5v-1.7c-2.8.6-3.4-1.3-3.4-1.3-.5-1.2-1.1-1.5-1.1-1.5-1-.6.1-.6.1-.6 1 .1 1.5 1.1 1.5 1.1.9 1.5 2.4 1.1 3 .8.1-.7.4-1.1.6-1.4-2.2-.3-4.5-1.1-4.5-5 0-1.1.4-2 1-2.7-.1-.3-.5-1.3.1-2.8 0 0 .9-.3 2.8 1a9.7 9.7 0 0 1 5 0c1.9-1.3 2.8-1 2.8-1 .6 1.5.2 2.5.1 2.8.6.7 1 1.6 1 2.7 0 3.9-2.3 4.7-4.5 5 .4.3.7.9.7 1.8v2.7c0 .3.2.6.7.5A10 10 0 0 0 12 2z" />
    </svg>
  ),
  JetBrains: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <rect x="3" y="3" width="18" height="18" rx="2" />
      <path d="M7 17h5M9 7v6" />
    </svg>
  ),
  Apple: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M17.05 12.5c0-2.7 2.2-4 2.3-4-1.3-1.8-3.2-2-3.9-2-1.6-.2-3.2 1-4 1-.8 0-2.1-1-3.4-1-1.8 0-3.4 1-4.3 2.6-1.8 3.2-.5 7.9 1.3 10.5.9 1.3 1.9 2.7 3.3 2.6 1.3-.05 1.8-.85 3.4-.85s2 .85 3.4.85c1.4 0 2.3-1.3 3.2-2.6 1-1.5 1.4-3 1.4-3-.1 0-2.7-1.05-2.7-4.1zM14.6 4.7c.7-.85 1.2-2.05 1.05-3.25-1 0-2.3.7-3 1.55-.65.75-1.3 2-1.1 3.15 1.15.1 2.35-.6 3.05-1.45z" />
    </svg>
  ),
  Windows: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M3 5.5 11 4.4v7.1H3V5.5zm0 13L11 19.6v-7H3v6zm9 1.3 9 1.2v-8.5h-9v7.3zm0-15.5v7.2h9V3l-9 1.3z" />
    </svg>
  ),
};

export function Integrations() {
  const { t } = useTranslation("landing");
  return (
    <section id="integrations" className="py-[120px] px-4 sm:px-8 relative z-[2]">
      <div className="mx-auto max-w-[1200px]">
        <div className="flex flex-col gap-5 max-w-[720px]">
          <span className="eyebrow">{t("integrations.eyebrow")}</span>
          <h2 className="font-display font-medium text-[clamp(1.875rem,4.4vw,3.5rem)] leading-[1.05] tracking-[-0.03em] text-balance">
            {t("integrations.title")}
          </h2>
          <p className="text-muted-foreground max-w-[56ch] text-pretty text-[clamp(0.9375rem,1.4vw,1.125rem)]">
            {t("integrations.sub")}
          </p>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3 mt-10">
          {ITEMS.map((it) => (
            <div
              key={it.name}
              className="aspect-[1.2/1] rounded-xl border flex flex-col items-center justify-center gap-2.5 text-[12px] font-mono text-muted-foreground transition-all hover:text-foreground relative overflow-hidden"
              style={{
                borderColor: "hsl(var(--border))",
                background: "rgba(255,255,255,0.01)",
              }}
            >
              {it.ready && (
                <span
                  className="absolute top-2 right-2 h-1.5 w-1.5 rounded-full"
                  style={{
                    background: "hsl(var(--success))",
                    boxShadow: "0 0 8px hsl(var(--success))",
                  }}
                />
              )}
              {it.soon && (
                <span
                  className="absolute top-2 right-2 text-[9px] px-1.5 py-0.5 rounded"
                  style={{
                    color: "hsl(var(--muted-foreground))",
                    border: "1px solid hsl(var(--border))",
                  }}
                >
                  v1
                </span>
              )}
              <div className="h-7 w-7">{Icons[it.icon]}</div>
              <span>{it.name}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
