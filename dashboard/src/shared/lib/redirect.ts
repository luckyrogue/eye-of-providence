// Helpers для preserve-destination-across-login flow.
//
// Когда unauth-redirect отправляет юзера на /login, текущий path сохраняется
// в `?redirect=`. После успешного login AuthRoute читает param и navigate'ит
// туда вместо дефолта.
//
// Validation: path должен быть локальным (`/...`, не `//evil.com`,
// не `https://...`) и не auth-page (защита от loop'а).

const AUTH_PAGES = new Set(["/login", "/signup", "/forgot-password", "/reset-password"]);

const REDIRECT_PARAM = "redirect";

// safeRedirect — нормализует path. Возвращает null если значение
// нерелевантно (cross-origin, auth-page, malformed).
export function safeRedirect(value: string | null | undefined): string | null {
  if (!value) return null;
  if (!value.startsWith("/")) return null;
  // Защита от //evil.com — protocol-relative URL'ы.
  if (value.startsWith("//")) return null;
  // Защита от data:/javascript: тоже (на случай если проверки выше обойдут).
  if (value.includes(":")) return null;
  // Anti-loop: redirect на auth-страницу теряет смысл.
  const pathOnly = value.split("?")[0].split("#")[0];
  if (AUTH_PAGES.has(pathOnly)) return null;
  return value;
}

// buildLoginURL — формирует /login?redirect=<encoded path> или /login
// если path пустой / unsafe.
export function buildLoginURL(currentPath: string): string {
  const safe = safeRedirect(currentPath);
  if (!safe) return "/login";
  return `/login?${REDIRECT_PARAM}=${encodeURIComponent(safe)}`;
}

// readRedirect — извлекает и validates param из URLSearchParams.
export function readRedirect(params: URLSearchParams): string | null {
  return safeRedirect(params.get(REDIRECT_PARAM));
}
