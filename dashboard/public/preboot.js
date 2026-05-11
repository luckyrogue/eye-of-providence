// preboot.js — sync-loaded ДО React mount'а. Закрывает FOUC и i18n flicker:
// 1) <html lang>=detected locale (синхронно с LOCALE_STORAGE_KEY из @eop/i18n);
// 2) data-theme="eop" (warm-dark + orange палитра из ui/styles.css).
// Вынесен из inline <script>, чтобы CSP script-src мог быть 'self' без
// 'unsafe-inline' (anti-XSS hardening).

(function () {
  try {
    var supported = ["ru", "en", "kk", "es"];
    var raw = localStorage.getItem("eop_locale");
    var lng = raw && supported.indexOf(raw) !== -1 ? raw : null;
    if (!lng && navigator.language) {
      var base = navigator.language.split("-")[0];
      if (supported.indexOf(base) !== -1) lng = base;
    }
    document.documentElement.lang = lng || "ru";
    document.documentElement.setAttribute("data-theme", "eop");
  } catch (_) {
    /* localStorage может быть недоступен в iframe/private mode */
  }
})();
