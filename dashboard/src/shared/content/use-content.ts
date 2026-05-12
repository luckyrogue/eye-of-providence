// useContent — Phase 4 CMS-lite read hook.
//
// Contract:
//   useContent<T>(slug, fallback) => T
// Behaviour:
//   1. Resolves current locale via i18next (default "ru").
//   2. Fires `GET /v1/content/:slug?locale=<l>` through react-query.
//      - retry: false  — a missing slug or DB outage should fall back
//        to bundled i18n immediately, not retry-storm the public API.
//      - staleTime aligned with backend `s-maxage=600` so re-renders
//        of the same slug don't refetch within the cache window.
//   3. If `?preview=<slug>` is in the URL AND this hook's slug matches,
//      uses `usePreviewContent` to overlay the draft for that one slug.
//      Other slugs continue to render published content as normal.
//   4. On error / 404 / empty content, returns the bundled `fallback`.
//
// The `fallback` argument MUST be a fully-shaped object — never null.
// The page can never render blank: synchronous fallback first, async
// CMS overlay later when react-query settles.

import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { http } from "../api/http";
import { usePreviewContext } from "./preview-context";
import type { ContentLocale, ContentResponse, ContentSlug } from "./types";

const STALE_MS = 5 * 60 * 1000; // 5 min — backend Cache-Control max-age=300.

function isSupportedLocale(l: string): l is ContentLocale {
  return l === "en" || l === "ru" || l === "kk" || l === "es";
}

export function useContentLocale(): ContentLocale {
  const { i18n } = useTranslation();
  const raw = i18n.resolvedLanguage ?? i18n.language ?? "ru";
  const base = raw.split("-")[0]?.toLowerCase() ?? "ru";
  return isSupportedLocale(base) ? base : "ru";
}

// usePublishedContent — published-only read (public endpoint).
function usePublishedContent<T>(slug: ContentSlug, locale: ContentLocale, enabled: boolean) {
  return useQuery({
    queryKey: ["content", slug, locale],
    enabled,
    staleTime: STALE_MS,
    gcTime: STALE_MS * 2,
    retry: false,
    refetchOnWindowFocus: false,
    queryFn: async () => {
      const r = await http.get<ContentResponse<T>>(`/v1/content/${encodeURIComponent(slug)}`, {
        params: { locale },
      });
      return r.data.content;
    },
  });
}

// usePreviewContent — draft-aware admin read for super_admins inside the
// preview overlay. Failures fall through to the published value silently.
function usePreviewContent<T>(slug: ContentSlug, locale: ContentLocale, enabled: boolean) {
  return useQuery({
    queryKey: ["content", "preview", slug, locale],
    enabled,
    staleTime: 0,
    retry: false,
    refetchOnWindowFocus: false,
    queryFn: async () => {
      const r = await http.get<{
        slug: string;
        locale: string;
        content: T;
        draft_content?: T | null;
      }>(`/v1/admin/content/${encodeURIComponent(slug)}`, {
        params: { locale, include_draft: true },
      });
      return r.data.draft_content ?? r.data.content;
    },
  });
}

// Empty-content guard — backend may return `{}` or `null` for a slug that
// was reverted mid-flight; treat as "no content" and fall back.
function isEmpty(value: unknown): boolean {
  if (value == null) return true;
  if (typeof value === "object" && !Array.isArray(value)) {
    return Object.keys(value as Record<string, unknown>).length === 0;
  }
  return false;
}

export function useContent<T>(slug: ContentSlug, fallback: T): T {
  const locale = useContentLocale();
  const preview = usePreviewContext();
  const isPreviewingThis = preview.previewSlug === slug;

  const published = usePublishedContent<T>(slug, locale, !isPreviewingThis);
  const draft = usePreviewContent<T>(slug, locale, isPreviewingThis);

  const value = isPreviewingThis ? draft.data : published.data;
  if (value === undefined || value === null || isEmpty(value)) return fallback;
  return value;
}
