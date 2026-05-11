import { useTranslation } from "react-i18next";
import { Eye } from "lucide-react";

export function Nav() {
  const { t } = useTranslation("landing");
  const isAuthed = typeof window !== "undefined" && !!localStorage.getItem("eop_user_id");
  const navLinks = [
    { href: "#features", label: t("nav.features") },
    { href: "#how", label: t("nav.how") },
    { href: "#pricing", label: t("nav.pricing") },
    { href: "#faq", label: t("nav.faq") },
  ];
  return (
    <header
      className="fixed top-0 left-0 right-0 z-50 backdrop-blur-md bg-background/60 border-b"
      style={{ borderColor: "hsl(var(--border))" }}
    >
      <div className="mx-auto max-w-7xl px-4 sm:px-8 h-[68px] flex items-center justify-between">
        <a href="/" className="flex items-center gap-3 group min-w-0">
          <div className="h-7 w-7 shrink-0 grid place-items-center relative">
            <Eye className="h-5 w-5 text-foreground transition-transform duration-300 group-hover:rotate-[8deg]" />
          </div>
          <span className="font-display font-semibold tracking-tight text-[15px] truncate">
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
        <div className="flex items-center gap-3">
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
            className="btn-eop-primary inline-flex items-center gap-2 px-4 py-2 rounded-lg text-[14px] font-medium whitespace-nowrap"
          >
            {isAuthed ? t("nav.open_dashboard") : t("nav.get_started")}
          </a>
        </div>
      </div>
    </header>
  );
}
