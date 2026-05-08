import { Button } from "@eop/ui";
import { Eye } from "lucide-react";

const NAV_LINKS = [
  { href: "#features", label: "Features" },
  { href: "#how", label: "How it works" },
  { href: "#pricing", label: "Pricing" },
  { href: "#faq", label: "FAQ" },
];

export function Nav() {
  const isAuthed = typeof window !== "undefined" && !!localStorage.getItem("eop_user_id");
  return (
    <header className="sticky top-0 z-40 border-b header-blur">
      <div className="mx-auto max-w-6xl px-6 h-16 flex items-center justify-between">
        <a href="/" className="flex items-center gap-2.5 group">
          <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center transition-transform duration-300 ease-out-expo group-hover:rotate-[8deg]">
            <Eye className="h-4 w-4 text-primary-foreground" />
          </div>
          <span className="font-display font-bold tracking-tightest text-lg">Eye of Providence</span>
        </a>
        <nav className="hidden md:flex items-center gap-8 text-sm">
          {NAV_LINKS.map((l) => (
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
              Sign in
            </a>
          )}
          <Button asChild size="sm">
            <a href="/dashboard">{isAuthed ? "Open dashboard" : "Get started"}</a>
          </Button>
        </div>
      </div>
    </header>
  );
}
