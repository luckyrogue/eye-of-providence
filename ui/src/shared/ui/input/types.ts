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
  autoCapitalize?: string;
  spellCheck?: boolean;
  maxLength?: number;
  minLength?: number;
  inputMode?: "text" | "tel" | "url" | "email" | "numeric" | "decimal" | "search" | "none";
  disabled?: boolean;
  className?: string;
  translateError?: (msg?: string) => string | undefined;
  hideMessage?: boolean;
  transform?: (raw: string) => string;
};
