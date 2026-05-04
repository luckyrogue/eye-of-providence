// Timezone preference: хранится в localStorage, по умолчанию — браузерный default.

const KEY = "eop_tz";

export function getTz(): string {
  return localStorage.getItem(KEY) || Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
}

export function setTz(tz: string) {
  localStorage.setItem(KEY, tz);
}

export const TIMEZONES: { label: string; value: string }[] = [
  { label: "Москва (UTC+3)", value: "Europe/Moscow" },
  { label: "Санкт-Петербург (UTC+3)", value: "Europe/Moscow" },
  { label: "Казань (UTC+3)", value: "Europe/Moscow" },
  { label: "Алматы (UTC+5)", value: "Asia/Almaty" },
  { label: "Ташкент (UTC+5)", value: "Asia/Tashkent" },
  { label: "Бишкек (UTC+6)", value: "Asia/Bishkek" },
  { label: "Новосибирск (UTC+7)", value: "Asia/Novosibirsk" },
  { label: "Иркутск (UTC+8)", value: "Asia/Irkutsk" },
  { label: "Владивосток (UTC+10)", value: "Asia/Vladivostok" },
  { label: "Берлин (UTC+1)", value: "Europe/Berlin" },
  { label: "Лондон (UTC+0)", value: "Europe/London" },
  { label: "New York (UTC-5)", value: "America/New_York" },
  { label: "San Francisco (UTC-8)", value: "America/Los_Angeles" },
  { label: "UTC", value: "UTC" },
];

// Уникальные значения для select
export const UNIQUE_TIMEZONES = Array.from(
  new Map(TIMEZONES.map((t) => [t.value, t])).values(),
);

export function formatDate(iso: string, tz: string): string {
  return new Date(iso).toLocaleString("ru-RU", {
    timeZone: tz,
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatTime(iso: string, tz: string): string {
  return new Date(iso).toLocaleTimeString("ru-RU", {
    timeZone: tz,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}
