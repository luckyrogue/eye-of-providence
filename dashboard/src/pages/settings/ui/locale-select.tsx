import { useState } from "react";
import { useTranslation } from "react-i18next";
import { SimpleSelect } from "@eop/ui";
import { updateLocale } from "@/entities/user";
import { LOCALE_LABELS, LOCALE_STORAGE_KEY, SUPPORTED_LOCALES, type Locale } from "@/shared/i18n";

export function LocaleSelect() {
  const { i18n } = useTranslation();
  const [locale, setLocale] = useState<Locale>((i18n.resolvedLanguage as Locale) || "ru");

  function change(next: Locale) {
    setLocale(next);
    localStorage.setItem(LOCALE_STORAGE_KEY, next);
    void i18n.changeLanguage(next);
    // Best-effort persist на backend, чтобы emails (invite/reset/digest)
    // приходили на нужном языке. Игнорируем ошибку — locale уже сохранён локально.
    void updateLocale(next).catch(() => undefined);
  }

  return (
    <SimpleSelect
      value={locale}
      onValueChange={(v) => change(v as Locale)}
      triggerClassName="w-full"
      options={SUPPORTED_LOCALES.map((code) => ({ value: code, label: LOCALE_LABELS[code] }))}
    />
  );
}
