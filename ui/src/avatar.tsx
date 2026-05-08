import { cn } from "./cn";

export function Avatar({
  name,
  size = "md",
  className,
}: {
  name: string;
  size?: "sm" | "md" | "lg";
  className?: string;
}) {
  const sizes = {
    sm: "h-7 w-7 text-[10px]",
    md: "h-9 w-9 text-sm",
    lg: "h-12 w-12 text-base",
  };
  const initials = name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((s) => s[0]?.toUpperCase() ?? "")
    .join("") || "?";
  return (
    <div
      className={cn(
        "rounded-full bg-gradient-to-br from-primary/30 to-primary/10 flex items-center justify-center font-medium shrink-0",
        sizes[size],
        className,
      )}
    >
      {initials}
    </div>
  );
}
