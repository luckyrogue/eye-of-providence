import type { ReactNode } from "react";
import type { Control, FieldPath, FieldValues } from "react-hook-form";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "./form";
import { Checkbox } from "./checkbox";
import { cn } from "../lib/cn";
export type CheckboxFieldProps<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
> = {
  control: Control<TFieldValues>;
  name: TName;
  label: ReactNode;
  description?: ReactNode;
  disabled?: boolean;
  className?: string;
  hideMessage?: boolean;
};
export function CheckboxField<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
>({
  control,
  name,
  label,
  description,
  disabled,
  className,
  hideMessage,
}: CheckboxFieldProps<TFieldValues, TName>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem className={cn("flex flex-row items-start gap-2 space-y-0", className)}>
          <FormControl>
            <Checkbox
              checked={!!field.value}
              onCheckedChange={(v) => field.onChange(!!v)}
              disabled={disabled}
              ref={field.ref}
              onBlur={field.onBlur}
            />
          </FormControl>
          <div className="space-y-0.5 leading-none">
            <FormLabel className="text-sm font-medium cursor-pointer">{label}</FormLabel>
            {description && <p className="text-xs text-muted-foreground">{description}</p>}
            {!hideMessage && <FormMessage />}
          </div>
        </FormItem>
      )}
    />
  );
}
