// JSON Schemas per slug. Mirrors backend allowlist (.team/product-cms-slugs.md).
// Used by Monaco for autocomplete + live validation in the admin editor.
//
// Keep these in sync with backend `SlugRegistry` in
// backend/internal/content/registry.go. A schema bump on the BE requires
// updating this map AND the schema_version in the response.

import type { ContentSlug } from "./types";

// JSON Schema draft 2020-12 (Monaco understands draft-07+ for inline hints).
// We avoid ajv here — schemas are passed to Monaco as plain objects.
type JsonSchema = {
  $schema?: string;
  type?: string;
  required?: string[];
  properties?: Record<string, JsonSchema>;
  additionalProperties?: boolean | JsonSchema;
  minLength?: number;
  maxLength?: number;
  minItems?: number;
  maxItems?: number;
  items?: JsonSchema;
  pattern?: string;
  description?: string;
  default?: unknown;
};

const textBlock = (max: number): JsonSchema => ({
  type: "object",
  required: ["text"],
  additionalProperties: false,
  properties: {
    text: { type: "string", minLength: 1, maxLength: max },
  },
});

const ctaBlock: JsonSchema = {
  type: "object",
  required: ["label", "href"],
  additionalProperties: false,
  properties: {
    label: { type: "string", minLength: 1, maxLength: 40 },
    href: {
      type: "string",
      minLength: 1,
      maxLength: 500,
      pattern: "^(https://|/)",
      description: "Must start with https:// or /. javascript:, http://, mailto: rejected.",
    },
    external: { type: "boolean", default: false },
  },
};

const pricingTier: JsonSchema = {
  type: "object",
  required: ["tagline", "bullets"],
  additionalProperties: false,
  properties: {
    tagline: { type: "string", minLength: 1, maxLength: 80 },
    bullets: {
      type: "array",
      minItems: 3,
      maxItems: 8,
      items: { type: "string", minLength: 1, maxLength: 80 },
    },
  },
};

const faqItems: JsonSchema = {
  type: "object",
  required: ["items"],
  additionalProperties: false,
  properties: {
    items: {
      type: "array",
      minItems: 3,
      maxItems: 10,
      items: {
        type: "object",
        required: ["q", "a"],
        additionalProperties: false,
        properties: {
          q: { type: "string", minLength: 1, maxLength: 200 },
          a: { type: "string", minLength: 1, maxLength: 800 },
          anchor: {
            type: "string",
            minLength: 1,
            maxLength: 60,
            pattern: "^[a-z0-9]+(-[a-z0-9]+)*$",
          },
        },
      },
    },
  },
};

export const CONTENT_SCHEMAS: Record<ContentSlug, JsonSchema> = {
  "landing.hero.headline": textBlock(200),
  "landing.hero.subhead": textBlock(400),
  "landing.hero.cta_primary": ctaBlock,
  "landing.hero.cta_secondary": ctaBlock,
  "landing.pricing.free.description": pricingTier,
  "landing.pricing.pro.description": pricingTier,
  "landing.pricing.business.description": pricingTier,
  "landing.faq.items": faqItems,
};

// Editor placeholders — pre-fill empty rows with a valid shape so admins
// don't start from a blank `{}`.
export const CONTENT_EXAMPLES: Record<ContentSlug, unknown> = {
  "landing.hero.headline": { text: "Ship faster. <em>Know how</em>." },
  "landing.hero.subhead": {
    text: "Track what you wrote, what your AI wrote, and where the rest of your day went.",
  },
  "landing.hero.cta_primary": { label: "Start free trial", href: "/signup", external: false },
  "landing.hero.cta_secondary": { label: "See live demo", href: "/demo", external: false },
  "landing.pricing.free.description": {
    tagline: "Free forever",
    bullets: ["Up to 5 users", "30 days history", "Basic dashboard"],
  },
  "landing.pricing.pro.description": {
    tagline: "For growing teams",
    bullets: ["Up to 50 users", "365 days history", "AI insights"],
  },
  "landing.pricing.business.description": {
    tagline: "For teams that need control",
    bullets: ["Unlimited users", "SSO", "Audit log"],
  },
  "landing.faq.items": {
    items: [
      { q: "What does the agent collect?", a: "Only metadata." },
      { q: "When does free end?", a: "Free tier stays free forever." },
      { q: "Is my code uploaded?", a: "No. Local-first by design." },
    ],
  },
};
