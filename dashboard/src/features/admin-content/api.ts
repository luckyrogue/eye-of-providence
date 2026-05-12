// Admin CMS-lite API hooks (Phase 4).
//
// Wraps `/v1/admin/content/*` endpoints. Concurrent-edit safety lives
// here: PUT carries an `If-Match: "<etag>"` header (= last-known etag),
// and a 412 surfaces a typed `ConcurrentEditError` callers can localise.

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../shared/api/http";
import type { ContentLocale, ContentSlug } from "../../shared/content";

export type AdminContentListItem = {
  slug: ContentSlug;
  locale: ContentLocale;
  has_published: boolean;
  has_draft: boolean;
  updated_at: string | null;
  updated_by_email: string | null;
};

export type AdminContentDetail = {
  slug: ContentSlug;
  locale: ContentLocale;
  content: unknown | null;
  draft_content: unknown | null;
  schema_version: number;
  published_at: string | null;
  updated_at: string | null;
  updated_by: string | null;
  updated_by_email: string | null;
  etag: string | null;
};

export type ContentSavePayload = {
  content: unknown;
  publish: boolean;
};

export class ConcurrentEditError extends Error {
  currentEtag?: string;
  constructor(msg: string, currentEtag?: string) {
    super(msg);
    this.name = "ConcurrentEditError";
    this.currentEtag = currentEtag;
  }
}

const keys = {
  all: ["admin", "content"] as const,
  list: () => [...keys.all, "list"] as const,
  detail: (slug: ContentSlug, locale: ContentLocale) =>
    [...keys.all, "detail", slug, locale] as const,
};

async function fetchList(): Promise<AdminContentListItem[]> {
  const r = await http.get<{ items: AdminContentListItem[] }>("/v1/admin/content");
  return r.data.items ?? [];
}

async function fetchDetail(slug: ContentSlug, locale: ContentLocale): Promise<AdminContentDetail> {
  const r = await http.get<AdminContentDetail>(`/v1/admin/content/${encodeURIComponent(slug)}`, {
    params: { locale, include_draft: true },
  });
  return r.data;
}

async function saveDetail(
  slug: ContentSlug,
  locale: ContentLocale,
  payload: ContentSavePayload,
  etag: string | null,
): Promise<AdminContentDetail> {
  try {
    const r = await http.put<AdminContentDetail>(
      `/v1/admin/content/${encodeURIComponent(slug)}`,
      { locale, content: payload.content, publish: payload.publish },
      { headers: etag ? { "If-Match": etag } : undefined },
    );
    return r.data;
  } catch (e) {
    const err = e as { status?: number; response?: { data?: { current_etag?: string } } };
    if (err.status === 412) {
      throw new ConcurrentEditError(
        "Content was modified by someone else",
        err.response?.data?.current_etag,
      );
    }
    throw e;
  }
}

async function revertDetail(slug: ContentSlug, locale: ContentLocale): Promise<void> {
  await http.delete(`/v1/admin/content/${encodeURIComponent(slug)}`, { params: { locale } });
}

export const useContentList = () => useQuery({ queryKey: keys.list(), queryFn: fetchList });

export const useContentBlock = (slug: ContentSlug | null, locale: ContentLocale | null) =>
  useQuery({
    queryKey: slug && locale ? keys.detail(slug, locale) : [...keys.all, "disabled"],
    queryFn: () => fetchDetail(slug!, locale!),
    enabled: !!slug && !!locale,
  });

export function useSaveContent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (vars: {
      slug: ContentSlug;
      locale: ContentLocale;
      payload: ContentSavePayload;
      etag: string | null;
    }) => saveDetail(vars.slug, vars.locale, vars.payload, vars.etag),
    onSuccess: (_d, vars) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: keys.list() }),
        qc.invalidateQueries({ queryKey: keys.detail(vars.slug, vars.locale) }),
        // Public reads share the same slug/locale key — invalidate so the
        // landing page picks up the new published row without a hard reload.
        qc.invalidateQueries({ queryKey: ["content", vars.slug, vars.locale] }),
      ]),
  });
}

export function useRevertContent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (vars: { slug: ContentSlug; locale: ContentLocale }) =>
      revertDetail(vars.slug, vars.locale),
    onSuccess: (_d, vars) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: keys.list() }),
        qc.invalidateQueries({ queryKey: keys.detail(vars.slug, vars.locale) }),
        qc.invalidateQueries({ queryKey: ["content", vars.slug, vars.locale] }),
      ]),
  });
}
