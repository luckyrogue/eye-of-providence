import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
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
    <header className="sticky top-0 z-40 border-b header-blur">
      <div className="mx-auto max-w-6xl px-4 sm:px-6 h-16 flex items-center justify-between">
        <a href="/" className="flex items-center gap-2 sm:gap-2.5 group min-w-0">
          <div className="h-8 w-8 shrink-0 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center transition-transform duration-300 ease-out-expo group-hover:rotate-[8deg]">
            <Eye className="h-4 w-4 text-primary-foreground" />
          </div>
          <span className="font-display font-bold tracking-tightest text-base sm:text-lg truncate">
            <span className="sm:hidden">EoP</span>
            <span className="hidden sm:inline">Eye of Providence</span>
          </span>
        </a>
        <nav className="hidden md:flex items-center gap-8 text-sm">
          {navLinks.map((l) => (
            <a key={l.href} href={l.href} className="text-muted-foreground hover:text-foreground transition-colors">
              {l.label}
            </a>
          ))}
        </nav>
        <div className="flex items-center gap-2">
          {!isAuthed && (
            <a
              href="/dashboard"
              className="hidden sm:inline-flex items-center text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              {t("nav.sign_in")}
            </a>
          )}
          <Button asChild size="sm">
            <a href="/dashboard">{isAuthed ? t("nav.open_dashboard") : t("nav.get_started")}</a>
          </Button>
        </div>
      </div>
    </header>
  );
}
