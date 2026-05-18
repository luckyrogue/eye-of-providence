/** Логотип EOP (треугольник + зрачок). Цвет через `currentColor` — задайте `text-accent`. */
export function EopMark({ size = 26, className }: { size?: number; className?: string }) {
  return (
    <svg viewBox="0 0 28 28" width={size} height={size} className={className} aria-hidden>
      <polygon
        points="14,3 26,24 2,24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinejoin="round"
      />
      <circle cx="14" cy="17" r="3" fill="none" stroke="currentColor" strokeWidth="1.4" />
      <circle cx="14" cy="17" r="1" fill="currentColor" />
    </svg>
  );
}
