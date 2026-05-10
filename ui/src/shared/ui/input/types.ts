import type { HTMLInputTypeAttribute, ReactNode } from "react";
import type { Control, FieldPath, FieldValues } from "react-hook-form";

export type InputFieldProps<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
> = {
  control: Control<TFieldValues>;
  name: TName;
  label: ReactNode;
  type?: HTMLInputTypeAttribute;
  placeholder?: string;
  autoComplete?: string;
  disabled?: boolean;
  className?: string;
  /** Maps validation message (e.g. i18n key from zod) to display text. */
  translateError?: (msg?: string) => string | undefined;
  hideMessage?: boolean;
};
