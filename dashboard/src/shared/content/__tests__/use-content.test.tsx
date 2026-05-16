// Tests for useContent: Phase 4 CMS-lite read hook with graceful fallback.
//
// Three scenarios:
//   1. Server returns content   → hook surfaces the server payload.
//   2. Server errors (DB down)  → hook returns the bundled `fallback`.
//   3. Locale switch            → hook re-fetches under the new locale.
//
// http.get is mocked at the module level so we can assert call args
// without spinning up MSW.

import { describe, expect, it, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import { I18nextProvider } from "react-i18next";
import i18next from "i18next";
import { initReactI18next } from "react-i18next";
import { PreviewProvider } from "../preview-context";
import { useContent } from "../use-content";
import type { ContentSlug, TextBlock } from "../types";

// Module-level http mock — single shared spy per test file.
vi.mock("../../api/http", () => ({
  http: { get: vi.fn() },
}));

// Re-import after the mock so we get the typed handle.
import { http } from "../../api/http";

const getMock = http.get as unknown as ReturnType<typeof vi.fn>;

// Minimal i18next instance for tests: only what useContent reads
// (i18n.resolvedLanguage / i18n.language). No real resources loaded.
function makeI18n(lng: string) {
  const inst = i18next.createInstance();
  // initReactI18next is required so `useTranslation()` resolves an instance.
  void inst.use(initReactI18next).init({
    lng,
    fallbackLng: "ru",
    resources: { [lng]: { common: {} } },
    react: { useSuspense: false },
  });
  return inst;
}

function wrapperFor(lng: string) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  const i18n = makeI18n(lng);
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={qc}>
        <I18nextProvider i18n={i18n}>
          <MemoryRouter initialEntries={["/"]}>
            <PreviewProvider>{children}</PreviewProvider>
          </MemoryRouter>
        </I18nextProvider>
      </QueryClientProvider>
    );
  };
}

const slug: ContentSlug = "landing.hero.headline";
const fallback: TextBlock = { text: "Bundled default." };

describe("useContent", () => {
  beforeEach(() => {
    getMock.mockReset();
  });

  it("returns the bundled fallback when the fetch errors (DB outage)", async () => {
    getMock.mockRejectedValueOnce(new Error("upstream unreachable"));
    const { result } = renderHook(() => useContent<TextBlock>(slug, fallback), {
      wrapper: wrapperFor("en"),
    });
    // First render is synchronous — fallback already there.
    expect(result.current).toEqual(fallback);
    // Even after the query settles, we still see the fallback.
    await waitFor(() => expect(getMock).toHaveBeenCalledTimes(1));
    expect(result.current).toEqual(fallback);
  });

  it("returns the server content when the fetch resolves", async () => {
    const server: TextBlock = { text: "From CMS!" };
    getMock.mockResolvedValueOnce({
      data: { slug, locale: "en", content: server, schema_version: 1 },
    });
    const { result } = renderHook(() => useContent<TextBlock>(slug, fallback), {
      wrapper: wrapperFor("en"),
    });
    await waitFor(() => expect(result.current).toEqual(server));
  });

  it("refetches when the locale changes", async () => {
    const en: TextBlock = { text: "EN copy" };
    const ru: TextBlock = { text: "RU copy" };
    getMock.mockImplementation(((_url: string, init?: { params?: { locale?: string } }) => {
      const locale = init?.params?.locale;
      return Promise.resolve({
        data: { slug, locale, content: locale === "ru" ? ru : en, schema_version: 1 },
      });
    }) as never);

    // Custom wrapper so we can flip the locale on the same i18n instance.
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 } } });
    const i18n = makeI18n("en");
    const Wrap = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={qc}>
        <I18nextProvider i18n={i18n}>
          <MemoryRouter initialEntries={["/"]}>
            <PreviewProvider>{children}</PreviewProvider>
          </MemoryRouter>
        </I18nextProvider>
      </QueryClientProvider>
    );

    const { result } = renderHook(() => useContent<TextBlock>(slug, fallback), { wrapper: Wrap });
    await waitFor(() => expect(result.current).toEqual(en));

    await act(async () => {
      await i18n.changeLanguage("ru");
    });

    await waitFor(() => expect(result.current).toEqual(ru));
    // Two distinct fetches — one per locale.
    const localesCalled = getMock.mock.calls.map(
      (c) => (c[1] as { params?: { locale?: string } } | undefined)?.params?.locale,
    );
    expect(localesCalled).toContain("en");
    expect(localesCalled).toContain("ru");
  });
});
