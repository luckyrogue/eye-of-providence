import { Eye } from "lucide-react";

const PRODUCT_LINKS = [
  { href: "#features", label: "Features" },
  { href: "#pricing", label: "Pricing" },
  { href: "#faq", label: "FAQ" },
  { href: "/dashboard", label: "Dashboard" },
];

const LEGAL_LINKS = [
  { href: "/privacy", label: "Privacy" },
  { href: "/terms", label: "Terms" },
  { href: "/security", label: "Security" },
];

export function Footer() {
  return (
    <footer className="border-t">
      <div className="mx-auto max-w-6xl px-6 py-12 grid grid-cols-2 md:grid-cols-4 gap-8 text-sm">
        <div className="col-span-2">
          <div className="flex items-center gap-2.5">
            <div className="h-7 w-7 rounded-md bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
              <Eye className="h-3.5 w-3.5 text-primary-foreground" />
            </div>
            <span className="font-display font-bold tracking-tightest">Eye of Providence</span>
          </div>
          <p className="text-muted-foreground mt-3 max-w-xs leading-relaxed text-xs">
            Privacy-first AI workflow analytics for engineers. Self-hostable.
          </p>
        </div>
        <FooterColumn title="Product" links={PRODUCT_LINKS} />
        <FooterColumn title="Legal" links={LEGAL_LINKS} />
      </div>
      <div className="border-t">
        <div className="mx-auto max-w-6xl px-6 py-5 flex items-center justify-between text-xs text-muted-foreground font-mono">
          <span>© {new Date().getFullYear()} Eye of Providence</span>
          <span className="flex items-center gap-1.5">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
            All systems operational
          </span>
        </div>
      </div>
    </footer>
  );
}

function FooterColumn({ title, links }: { title: string; links: { href: string; label: string }[] }) {
  return (
    <div>
      <div className="font-mono text-[11px] uppercase tracking-widest2 text-muted-foreground mb-3">{title}</div>
      <ul className="space-y-2 text-muted-foreground">
        {links.map((l) => (
          <li key={l.href}>
            <a href={l.href} className="hover:text-foreground transition-colors">
              {l.label}
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
}
