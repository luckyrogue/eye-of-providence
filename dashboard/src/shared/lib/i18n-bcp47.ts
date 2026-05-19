/** i18n storage ("ru"/"en"/…) → BCP-47 для Intl. */
export function i18nToBcp47(lng: string): string {
  const base = lng.split("-")[0];
  switch (base) {
    case "ru":
      return "ru-RU";
    case "kk":
      return "kk-KZ";
    case "es":
      return "es-ES";
    case "en":
    default:
      return "en-US";
  }
}
