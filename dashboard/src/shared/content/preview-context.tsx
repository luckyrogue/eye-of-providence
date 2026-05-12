// PreviewContext — exposes the single `?preview=<slug>` value to landing
// children so each useContent() can decide independently whether to overlay
// its slug with draft content.
//
// Phase 1 explicitly limits previews to one slug per session (see
// product-phase4-brief.md "Out of scope"). Multi-slug preview waits for
// Phase 5.

import { createContext, useContext, useMemo, type ReactNode } from "react";
import { useSearchParams } from "react-router-dom";
import type { ContentSlug } from "./types";
import { CONTENT_SLUGS } from "./types";

type PreviewState = {
  previewSlug: ContentSlug | null;
};

const PreviewCtx = createContext<PreviewState>({ previewSlug: null });

function isContentSlug(s: string): s is ContentSlug {
  return (CONTENT_SLUGS as string[]).includes(s);
}

export function PreviewProvider({ children }: { children: ReactNode }) {
  const [params] = useSearchParams();
  const raw = params.get("preview");
  const value = useMemo<PreviewState>(() => {
    if (raw && isContentSlug(raw)) return { previewSlug: raw };
    return { previewSlug: null };
  }, [raw]);
  return <PreviewCtx.Provider value={value}>{children}</PreviewCtx.Provider>;
}

export function usePreviewContext(): PreviewState {
  return useContext(PreviewCtx);
}
