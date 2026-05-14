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
