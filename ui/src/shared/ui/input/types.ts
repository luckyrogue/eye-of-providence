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
  /** Passed through to the native `<input>`. */
  maxLength?: number;
  minLength?: number;
  inputMode?: "text" | "tel" | "url" | "email" | "numeric" | "decimal" | "search" | "none";
  disabled?: boolean;
  className?: string;
  /** Maps validation message (e.g. i18n key from zod) to display text. */
  translateError?: (msg?: string) => string | undefined;
  hideMessage?: boolean;
  /** Normalize user input before storing in form state (e.g. strip spaces, uppercase). */
  transform?: (raw: string) => string;
};
