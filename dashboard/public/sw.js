// Eye of Providence service worker — offline shell + future push hook.
//
// Caching strategy:
//   - Navigation (HTML): network-first → fallback на cached index.html.
//     Гарантирует, что юзер увидит latest UI online, и shell offline.
//   - /assets/* (Vite hashed bundles): cache-first immutable.
//   - /v1/*: NEVER cached — user data, PII, must hit network.
//   - Static (favicon, manifest, icons): cache-first revalidate via etag.
//
// Push event handler заглушка готова под Web Push API (Phase B).

const VERSION = "v2";
const SHELL_CACHE = `eop-shell-${VERSION}`;
const ASSET_CACHE = `eop-assets-${VERSION}`;
const SHELL_URLS = ["/", "/index.html", "/manifest.webmanifest", "/favicon.svg"];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(SHELL_CACHE).then((cache) =>
      // addAll fail-fast: если хоть один URL не достижим — install отказан.
      cache.addAll(SHELL_URLS).catch(() => {
        // best-effort, не блокируем install при temporary network issues
      }),
    ),
  );
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    (async () => {
      // Чистим старые версии кэша (у нас новые имена при VERSION bump).
      const keys = await caches.keys();
      await Promise.all(
        keys
          .filter((k) => k.startsWith("eop-") && !k.endsWith(VERSION))
          .map((k) => caches.delete(k)),
      );
      await self.clients.claim();
    })(),
  );
});

self.addEventListener("fetch", (event) => {
  const req = event.request;
  const url = new URL(req.url);

  // Только same-origin GET — POST/PUT/DELETE не cache'им, cross-origin тоже.
  if (req.method !== "GET" || url.origin !== self.location.origin) return;

  // API requests — bypass cache. Auth-sensitive, user-specific.
  if (url.pathname.startsWith("/v1/") || url.pathname.startsWith("/api/")) return;

  // Navigation — network-first с offline-fallback.
  if (req.mode === "navigate") {
    event.respondWith(
      (async () => {
        try {
          const fresh = await fetch(req);
          const cache = await caches.open(SHELL_CACHE);
          cache.put("/index.html", fresh.clone());
          return fresh;
        } catch {
          const cache = await caches.open(SHELL_CACHE);
          const cached = await cache.match("/index.html");
          if (cached) return cached;
          // Полный offline + нет cache → 504.
          return new Response("Offline", { status: 504 });
        }
      })(),
    );
    return;
  }

  // Locale markdown — never cache (updates without content hash; SW stale cache
  // caused old /legal/privacy.md to stick after deploy).
  if (url.pathname.startsWith("/legal/") || url.pathname.startsWith("/docs/")) {
    event.respondWith(fetch(req, { cache: "no-store" }));
    return;
  }

  // Hashed assets (Vite) — cache-first.
  if (url.pathname.startsWith("/assets/")) {
    event.respondWith(
      (async () => {
        const cache = await caches.open(ASSET_CACHE);
        const cached = await cache.match(req);
        if (cached) return cached;
        try {
          const fresh = await fetch(req);
          if (fresh.ok) cache.put(req, fresh.clone());
          return fresh;
        } catch {
          return new Response("Offline asset", { status: 504 });
        }
      })(),
    );
    return;
  }

  // Static (favicon/manifest/icons/changelog.json) — stale-while-revalidate.
  event.respondWith(
    (async () => {
      const cache = await caches.open(SHELL_CACHE);
      const cached = await cache.match(req);
      const network = fetch(req)
        .then((res) => {
          if (res.ok) cache.put(req, res.clone());
          return res;
        })
        .catch(() => cached);
      return cached || network;
    })(),
  );
});

//
// Когда server шлёт push-payload (после weekly report generated, anomaly
// alert и т.д.), browser вызывает это. Мы рендерим OS notification.
self.addEventListener("push", (event) => {
  let data = {};
  try {
    data = event.data ? event.data.json() : {};
  } catch {
    data = { title: "Eye of Providence", body: event.data?.text() ?? "" };
  }
  const title = data.title || "Eye of Providence";
  const options = {
    body: data.body || "",
    icon: "/icon-192.png",
    badge: "/icon-192.png",
    data: { url: data.url || "/dashboard" },
    tag: data.tag, // dedupes пушей с одинаковым tag
  };
  event.waitUntil(self.registration.showNotification(title, options));
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  const url = event.notification.data?.url || "/dashboard";
  event.waitUntil(
    self.clients.matchAll({ type: "window", includeUncontrolled: true }).then((clients) => {
      for (const c of clients) {
        if (c.url.includes(url) && "focus" in c) return c.focus();
      }
      if (self.clients.openWindow) return self.clients.openWindow(url);
    }),
  );
});
