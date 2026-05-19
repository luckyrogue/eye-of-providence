export type DocEntry = {
  slug: string;
  titleKey: string;
  path: string;
};

export const PUBLIC_DOCS: DocEntry[] = [
  { slug: "install", titleKey: "docs.install", path: "/docs/install.md" },
  { slug: "self-hosting", titleKey: "docs.self_hosting", path: "/docs/self-hosting.md" },
  { slug: "attribution", titleKey: "docs.attribution", path: "/docs/attribution.md" },
  { slug: "data-model", titleKey: "docs.data_model", path: "/docs/data-model.md" },
];

export type LegalSlug = "privacy" | "terms" | "security";

export const LEGAL_PAGES: Record<LegalSlug, { titleKey: string; path: string }> = {
  privacy: { titleKey: "legal.privacy", path: "/legal/privacy.md" },
  terms: { titleKey: "legal.terms", path: "/legal/terms.md" },
  security: { titleKey: "legal.security", path: "/legal/security.md" },
};
