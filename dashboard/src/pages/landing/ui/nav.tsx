import { useTranslation } from "react-i18next";
import { LangSwitch } from "../../../shared/ui/lang-switch";

export function Nav() {
  const { t } = useTranslation("landing");
  const isAuthed = typeof window !== "undefined" && !!localStorage.getItem("eop_user_id");
  const navLinks = [
    { href: "#measure", label: t("nav.features") },
    { href: "#attribution", label: t("nav.attribution") },
    { href: "#privacy", label: t("nav.privacy") },
    { href: "#integrations", label: t("nav.integrations") },
  ];
  return (
    <header
      className="fixed top-0 left-0 right-0 z-50 backdrop-blur-md bg-background/60 border-b"
      style={{ borderColor: "hsl(var(--border))" }}
    >
      <div className="mx-auto max-w-[1400px] px-4 sm:px-8 h-[68px] flex items-center justify-between">
        <a href="/" className="flex items-center gap-3 group min-w-0">
          <span className="h-7 w-7 grid place-items-center relative">
            <svg viewBox="0 0 28 28" width="28" height="28">
              <polygon
                points="14,3 26,24 2,24"
                fill="none"
                stroke="hsl(var(--accent))"
                strokeWidth="1.4"
              />
              <circle
                cx="14"
                cy="17"
                r="3"
                fill="none"
                stroke="hsl(var(--accent))"
                strokeWidth="1.4"
              />
              <circle cx="14" cy="17" r="1" fill="hsl(var(--accent))" />
            </svg>
          </span>
          <span className="font-display font-semibold tracking-[-0.01em] text-[15px] truncate">
            <span className="sm:hidden">EoP</span>
            <span className="hidden sm:inline">Eye of Providence</span>
          </span>
        </a>
        <nav className="hidden md:flex items-center gap-7 text-[14px]">
          {navLinks.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              {l.label}
            </a>
          ))}
        </nav>
        <div className="flex items-center gap-3.5">
          <LangSwitch />
          {!isAuthed && (
            <a
              href="/dashboard"
              className="hidden sm:inline-flex items-center text-[14px] text-muted-foreground hover:text-foreground transition-colors"
            >
              {t("nav.sign_in")}
            </a>
          )}
          <a
            href="/dashboard"
            className="btn-eop-primary inline-flex h-10 items-center gap-2 rounded-lg px-4 text-[14px] font-medium whitespace-nowrap"
          >
            {isAuthed ? t("nav.open_dashboard") : t("nav.get_started")}
          </a>
        </div>
      </div>
    </header>
  );
}
