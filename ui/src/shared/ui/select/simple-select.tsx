import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select-primitives";
import type { SimpleSelectOption } from "./types";

export type { SimpleSelectOption } from "./types";

export function SimpleSelect({
  value,
  onValueChange,
  options,
  placeholder,
  disabled,
  triggerClassName,
}: {
  value?: string;
  onValueChange?: (value: string) => void;
  options: SimpleSelectOption[];
  placeholder?: string;
  disabled?: boolean;
  triggerClassName?: string;
}) {
  return (
    <Select value={value} onValueChange={onValueChange} disabled={disabled}>
      <SelectTrigger className={triggerClassName}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {options.map((o) => (
          <SelectItem key={o.value} value={o.value} disabled={o.disabled}>
            {o.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
