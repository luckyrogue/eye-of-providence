// Низкоуровневая обёртка над localStorage для всех `eop_*` ключей.
// Без React, без entities — чтобы её мог использовать и `shared/api/http`,
// и `entities/session`, и `entities/user` без циклических импортов.

export const SESSION_KEYS = {
  userId: "eop_user_id",
  displayName: "eop_display_name",
  token: "eop_token",
  team: "eop_team",
} as const;

const PREFIX = "eop_";

export function getUserId(): string | null {
  return localStorage.getItem(SESSION_KEYS.userId);
}

export function getToken(): string | null {
  return localStorage.getItem(SESSION_KEYS.token);
}

export function setSession(payload: { user_id: string; token?: string; display_name?: string }) {
  localStorage.setItem(SESSION_KEYS.userId, payload.user_id);
  if (payload.token) localStorage.setItem(SESSION_KEYS.token, payload.token);
  if (payload.display_name) localStorage.setItem(SESSION_KEYS.displayName, payload.display_name);
}

// clearSession — снимает ВСЕ `eop_*` ключи. Используется как при явном
// logout, так и при 401-emit'е из http-interceptor.
export function clearSession() {
  try {
    for (let i = localStorage.length - 1; i >= 0; i--) {
      const k = localStorage.key(i);
      if (k && k.startsWith(PREFIX)) localStorage.removeItem(k);
    }
  } catch {
    /* private mode / disabled storage */
  }
}

// notifySessionChange — same-tab сигнал для useSyncExternalStore-подписчиков
// (cross-tab летит автоматически через native `storage` event).
export function notifySessionChange() {
  window.dispatchEvent(new Event("storage"));
}
