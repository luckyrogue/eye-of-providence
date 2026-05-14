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
export type TextBlock = {
  text: string;
};
export type CtaBlock = {
  label: string;
  href: string;
  external?: boolean;
};
export type PricingTierBlock = {
  tagline: string;
  bullets: string[];
};
export type FaqItem = {
  q: string;
  a: string;
  anchor?: string;
};
export type FaqItemsBlock = {
  items: FaqItem[];
};
export type ContentResponse<T> = {
  slug: string;
  locale: string;
  content: T;
  schema_version: number;
  published_at?: string;
  updated_at?: string;
};
