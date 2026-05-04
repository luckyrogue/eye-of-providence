// Content script — слушает copy events на сообщениях ассистента.
// Phase 2: per-domain селекторы, sha256 + size (без контента) -> background.

document.addEventListener("copy", () => {
  // Skeleton: реальная фильтрация (только если копируется из AI message bubble) — в Phase 2.
  console.debug("[eop] copy event on", location.host);
});
