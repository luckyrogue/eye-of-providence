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
  autoCapitalize,
  spellCheck,
  maxLength,
  minLength,
  inputMode,
  disabled,
  className,
  translateError,
  hideMessage,
  transform,
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
              autoCapitalize={autoCapitalize}
              spellCheck={spellCheck}
              maxLength={maxLength}
              minLength={minLength}
              inputMode={inputMode}
              disabled={disabled}
              className={cn(className)}
              {...field}
              onChange={(e) => {
                const raw = e.target.value;
                field.onChange(transform ? transform(raw) : raw);
              }}
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
