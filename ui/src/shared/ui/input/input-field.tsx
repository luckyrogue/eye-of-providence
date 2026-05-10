import type { FieldPath, FieldValues } from "react-hook-form";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "../form";
import { cn } from "../../lib/cn";
import { Input } from "./input";
import type { InputFieldProps } from "./types";

export type { InputFieldProps } from "./types";

/**
 * react-hook-form + shadcn FormField + Input (props instead of render-prop boilerplate).
 */
export function InputField<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
>({
  control,
  name,
  label,
  type = "text",
  placeholder,
  autoComplete,
  disabled,
  className,
  translateError,
  hideMessage,
}: InputFieldProps<TFieldValues, TName>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field, fieldState }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Input
              type={type}
              placeholder={placeholder}
              autoComplete={autoComplete}
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
