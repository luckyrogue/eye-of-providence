// shadcn/ui Badge — copy-paste из registry, расширено tonal-вариантами и
// `mono` для font-mono отображения.
import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "../lib/cn";

export const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium",
  {
    variants: {
      variant: {
        default: "bg-muted text-muted-foreground border-border",
        blue: "bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/30",
        purple: "bg-purple-500/10 text-purple-700 dark:text-purple-300 border-purple-500/30",
        amber: "bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30",
        green: "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30",
        red: "bg-destructive/10 text-destructive border-destructive/30",
      },
    },
    defaultVariants: { variant: "default" },
  },
);

export type BadgeVariant = VariantProps<typeof badgeVariants>["variant"];

interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {
  mono?: boolean;
}

export function Badge({ className, variant, mono, ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        badgeVariants({ variant }),
        mono && "font-mono uppercase tracking-widest2",
        className,
      )}
      {...props}
    />
  );
}
