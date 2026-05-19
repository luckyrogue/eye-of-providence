export type DocEntry = {
  slug: string;
  titleKey: string;
  /** Filename under `/docs/{locale}/`. */
  file: string;
};

export const PUBLIC_DOCS: DocEntry[] = [
  { slug: "install", titleKey: "docs.install", file: "install.md" },
  { slug: "self-hosting", titleKey: "docs.self_hosting", file: "self-hosting.md" },
  { slug: "attribution", titleKey: "docs.attribution", file: "attribution.md" },
  { slug: "data-model", titleKey: "docs.data_model", file: "data-model.md" },
];

export type LegalSlug = "privacy" | "terms" | "security";

export const LEGAL_PAGES: Record<LegalSlug, { titleKey: string; file: string }> = {
  privacy: { titleKey: "legal.privacy", file: "privacy.md" },
  terms: { titleKey: "legal.terms", file: "terms.md" },
  security: { titleKey: "legal.security", file: "security.md" },
};
