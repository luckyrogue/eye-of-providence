// Lang switcher — globe icon + dropdown с 4 locales.
// Заменил inline pill (EN|RU) — на mobile pill был полностью скрыт, что
// делало смену языка невозможной. Globe-dropdown компактнее, работает
// одинаково на mobile и desktop.

import { useTranslation } from "react-i18next";
import { Globe, Check } from "lucide-react";
import { LOCALE_STORAGE_KEY } from "@eop/i18n";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@eop/ui";

const LANGS = [
  { code: "en", label: "English" },
  { code: "ru", label: "Русский" },
  { code: "kk", label: "Қазақша" },
  { code: "es", label: "Español" },
] as const;

export function LangSwitch() {
  const { i18n } = useTranslation();
  const current = i18n.language.split("-")[0];
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        aria-label="Change language"
        className="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-full border transition-colors hover:bg-foreground/5 font-mono text-[12px] text-muted-foreground"
        style={{ borderColor: "hsl(var(--eop-line-strong))" }}
      >
        <Globe className="h-3.5 w-3.5" />
        <span className="uppercase">{current}</span>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[160px]">
        {LANGS.map((l) => (
          <DropdownMenuItem
            key={l.code}
            onClick={() => {
              localStorage.setItem(LOCALE_STORAGE_KEY, l.code);
              void i18n.changeLanguage(l.code);
            }}
            className="flex items-center justify-between gap-3 cursor-pointer"
          >
            <span className="flex items-center gap-2">
              <span className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground w-5">
                {l.code}
              </span>
              <span className="text-[13px]">{l.label}</span>
            </span>
            {current === l.code && <Check className="h-3.5 w-3.5 text-[hsl(var(--accent))]" />}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
