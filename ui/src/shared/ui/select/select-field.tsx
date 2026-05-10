import type { ReactNode } from "react";
import type { Control, FieldPath, FieldValues } from "react-hook-form";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "../form";
import { SimpleSelect } from "./simple-select";
import type { SimpleSelectOption } from "./types";

export type SelectFieldProps<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
> = {
  control: Control<TFieldValues>;
  name: TName;
  label: ReactNode;
  options: SimpleSelectOption[];
  placeholder?: string;
  disabled?: boolean;
  triggerClassName?: string;
  translateError?: (msg?: string) => string | undefined;
  hideMessage?: boolean;
};

export function SelectField<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
>({
  control,
  name,
  label,
  options,
  placeholder,
  disabled,
  triggerClassName,
  translateError,
  hideMessage,
}: SelectFieldProps<TFieldValues, TName>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field, fieldState }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <SimpleSelect
              value={field.value ?? ""}
              onValueChange={field.onChange}
              options={options}
              placeholder={placeholder}
              disabled={disabled}
              triggerClassName={triggerClassName}
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
