// CMS content types — Phase 4 (CMS-lite) front-end view.
//
// 8 Phase 1 slugs (see .team/product-cms-slugs.md). Backend slug registry
// is the source of truth; the FE mirrors only what's needed to render +
// give Monaco hints.

export type ContentSlug =
  | "landing.hero.headline"
  | "landing.hero.subhead"
  | "landing.hero.cta_primary"
  | "landing.hero.cta_secondary"
  | "landing.pricing.free.description"
  | "landing.pricing.pro.description"
  | "landing.pricing.business.description"
  | "landing.faq.items";

export const CONTENT_SLUGS: ContentSlug[] = [
  "landing.hero.headline",
  "landing.hero.subhead",
  "landing.hero.cta_primary",
  "landing.hero.cta_secondary",
  "landing.pricing.free.description",
  "landing.pricing.pro.description",
  "landing.pricing.business.description",
  "landing.faq.items",
];

export type ContentLocale = "en" | "ru" | "kk" | "es";
export const CONTENT_LOCALES: ContentLocale[] = ["en", "ru", "kk", "es"];

// Shapes per slug. Plain text only (no HTML — renderer interprets <em>
// in headline as a special case before render; see hero.tsx).

export type TextBlock = { text: string };
export type CtaBlock = { label: string; href: string; external?: boolean };
export type PricingTierBlock = { tagline: string; bullets: string[] };
export type FaqItem = { q: string; a: string; anchor?: string };
export type FaqItemsBlock = { items: FaqItem[] };

// ContentResponse — what `GET /v1/content/:slug?locale=...` returns.
// Frontend only consumes `content`; the rest is metadata for dev tools.
export type ContentResponse<T> = {
  slug: string;
  locale: string;
  content: T;
  schema_version: number;
  published_at?: string;
  updated_at?: string;
};
