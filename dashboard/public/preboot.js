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
  } catch (_) {}
})();
