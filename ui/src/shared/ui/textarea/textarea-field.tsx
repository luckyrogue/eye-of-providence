import type { ReactNode } from "react";
import type { Control, FieldPath, FieldValues } from "react-hook-form";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "../form";
import { cn } from "../../lib/cn";
import { Textarea } from "./textarea";
export type TextareaFieldProps<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
> = {
  control: Control<TFieldValues>;
  name: TName;
  label: ReactNode;
  placeholder?: string;
  rows?: number;
  disabled?: boolean;
  className?: string;
  translateError?: (msg?: string) => string | undefined;
  hideMessage?: boolean;
};
export function TextareaField<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
>({
  control,
  name,
  label,
  placeholder,
  rows,
  disabled,
  className,
  translateError,
  hideMessage,
}: TextareaFieldProps<TFieldValues, TName>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field, fieldState }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Textarea
              placeholder={placeholder}
              rows={rows}
              disabled={disabled}
              className={cn(className)}
              {...field}
            />
          </FormControl>
          {!hideMessage && (
            <FormMessage>
              {translateError ? translateError(fieldState.error?.message) : undefined}
            </FormMessage>
          )}
        </FormItem>
      )}
    />
  );
}
