const AUTH_PAGES = new Set(["/login", "/signup", "/forgot-password", "/reset-password"]);
const REDIRECT_PARAM = "redirect";
export function safeRedirect(value: string | null | undefined): string | null {
  if (!value) return null;
  if (!value.startsWith("/")) return null;
  if (value.startsWith("//")) return null;
  if (value.includes(":")) return null;
  const pathOnly = value.split("?")[0].split("#")[0];
  if (AUTH_PAGES.has(pathOnly)) return null;
  return value;
}
export function buildLoginURL(currentPath: string): string {
  const safe = safeRedirect(currentPath);
  if (!safe) return "/login";
  return `/login?${REDIRECT_PARAM}=${encodeURIComponent(safe)}`;
}
export function readRedirect(params: URLSearchParams): string | null {
  return safeRedirect(params.get(REDIRECT_PARAM));
}
