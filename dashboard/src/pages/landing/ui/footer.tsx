import { useTranslation } from "react-i18next";
import { Eye } from "lucide-react";
export function Footer() {
  const { t } = useTranslation("landing");
  const productLinks = [
    { href: "#features", label: t("nav.features") },
    { href: "#pricing", label: t("nav.pricing") },
    { href: "#faq", label: t("nav.faq") },
    { href: "/changelog", label: t("nav.changelog") },
    { href: "/dashboard", label: t("nav.open_dashboard") },
  ];
  const legalLinks = [
    { href: "/privacy", label: t("footer.legal_privacy") },
    { href: "/terms", label: t("footer.legal_terms") },
    { href: "/security", label: t("footer.legal_security") },
  ];
  return (
    <footer className="border-t px-4 sm:px-8 pt-16 pb-10">
      <div className="mx-auto max-w-7xl grid grid-cols-2 md:grid-cols-4 gap-10 text-[13px]">
        <div className="col-span-2 md:col-span-2">
          <div className="flex items-center gap-3">
            <Eye className="h-5 w-5 text-foreground" />
            <span className="font-display font-semibold tracking-tight text-[15px]">
              Eye of Providence
            </span>
          </div>
          <p className="text-muted-foreground mt-4 max-w-sm leading-relaxed text-[13px]">
            {t("footer.tagline")}
          </p>
        </div>
        <FooterColumn title={t("footer.product_title")} links={productLinks} />
        <FooterColumn title={t("footer.legal_title")} links={legalLinks} />
      </div>
      <div
        className="mx-auto max-w-7xl mt-10 pt-6 border-t flex flex-wrap items-center justify-between gap-3 text-[12px] text-muted-foreground font-mono"
        style={{ borderColor: "hsl(var(--border))" }}
      >
        <span>© {new Date().getFullYear()} Eye of Providence · MIT</span>
        <span className="flex items-center gap-1.5">
          <span
            className="h-1.5 w-1.5 rounded-full animate-pulse"
            style={{ background: "hsl(var(--success))" }}
          />
          {t("footer.status_ok")}
        </span>
      </div>
    </footer>
  );
}
function FooterColumn({
  title,
  links,
}: {
  title: string;
  links: {
    href: string;
    label: string;
  }[];
}) {
  return (
    <div>
      <h6 className="font-mono text-[11px] uppercase tracking-widest2 text-muted-foreground mb-4 font-medium">
        {title}
      </h6>
      <ul className="space-y-2.5 text-muted-foreground">
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
