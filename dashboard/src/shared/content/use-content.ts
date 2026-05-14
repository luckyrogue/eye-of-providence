import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { http } from "../api/http";
import { usePreviewContext } from "./preview-context";
import type { ContentLocale, ContentResponse, ContentSlug } from "./types";
const STALE_MS = 5 * 60 * 1000;
function isSupportedLocale(l: string): l is ContentLocale {
  return l === "en" || l === "ru" || l === "kk" || l === "es";
}
export function useContentLocale(): ContentLocale {
  const { i18n } = useTranslation();
  const raw = i18n.resolvedLanguage ?? i18n.language ?? "ru";
  const base = raw.split("-")[0]?.toLowerCase() ?? "ru";
  return isSupportedLocale(base) ? base : "ru";
}
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
