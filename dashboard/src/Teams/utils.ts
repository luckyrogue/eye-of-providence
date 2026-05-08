// Локальные хелперы для отображения ролей и склонений участников.

export function translateRole(r: string): string {
  return { owner: "владелец", admin: "админ", member: "участник" }[r] ?? r;
}

export function plural(n: number): string {
  if (n % 10 === 1 && n % 100 !== 11) return "";
  if ([2, 3, 4].includes(n % 10) && ![12, 13, 14].includes(n % 100)) return "а";
  return "ов";
}
