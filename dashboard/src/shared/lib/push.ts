// Утилиты для Web Push: декодинг VAPID public-key (RFC 4648 §5 base64url) и
// эвристика короткого browser+platform лейбла из User-Agent.
// Чистые функции без React, без браузерного state.

// VAPID public key приходит base64url-encoded. PushManager.subscribe требует
// Uint8Array. Padding восстанавливаем до кратности 4.
export function urlBase64ToUint8Array(b64: string): Uint8Array {
  const padding = "=".repeat((4 - (b64.length % 4)) % 4);
  const base64 = (b64 + padding).replace(/-/g, "+").replace(/_/g, "/");
  const rawData = window.atob(base64);
  const out = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; i++) out[i] = rawData.charCodeAt(i);
  return out;
}

// Короткий browser+platform лейбл из UA. Heuristic, не идеально, но хватает
// для UI ("список зарегистрированных устройств").
export function parseUserAgent(ua: string): string {
  if (!ua) return "";
  const platform = /iPhone|iPad/.test(ua)
    ? "iOS"
    : /Android/.test(ua)
      ? "Android"
      : /Macintosh/.test(ua)
        ? "Mac"
        : /Windows/.test(ua)
          ? "Windows"
          : /Linux/.test(ua)
            ? "Linux"
            : "";
  const browser = /Edg\//.test(ua)
    ? "Edge"
    : /Chrome\//.test(ua)
      ? "Chrome"
      : /Firefox\//.test(ua)
        ? "Firefox"
        : /Safari\//.test(ua)
          ? "Safari"
          : "";
  return [browser, platform].filter(Boolean).join(" • ");
}
