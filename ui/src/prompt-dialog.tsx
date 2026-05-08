import { useEffect, useState, type ReactNode } from "react";
import { Button } from "./button";
import { Input } from "./input";
import { Modal } from "./modal";

// PromptDialog — стилизованная замена window.prompt(): модалка с одним
// текстовым полем + кнопками. Возвращает строку или null (cancel).

export function PromptDialog({
  open,
  title,
  description,
  label,
  placeholder,
  initialValue = "",
  confirmText = "OK",
  cancelText = "Отмена",
  onClose,
  onConfirm,
  busy,
}: {
  open: boolean;
  title: ReactNode;
  description?: ReactNode;
  label?: string;
  placeholder?: string;
  initialValue?: string;
  confirmText?: string;
  cancelText?: string;
  onClose: () => void;
  onConfirm: (value: string) => void;
  busy?: boolean;
}) {
  const [value, setValue] = useState(initialValue);

  // Сбрасываем значение при каждом открытии модалки.
  useEffect(() => {
    if (open) setValue(initialValue);
  }, [open, initialValue]);

  function submit() {
    if (busy || !value.trim()) return;
    onConfirm(value.trim());
  }

  return (
    <Modal open={open} onClose={onClose} className="max-w-md">
      <div className="p-6 space-y-4">
        <h3 className="font-display text-xl tracking-tight">{title}</h3>
        {description && <div className="text-sm text-muted-foreground">{description}</div>}
        <Input
          autoFocus
          label={label}
          placeholder={placeholder}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && submit()}
        />
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="outline" onClick={onClose} disabled={busy}>
            {cancelText}
          </Button>
          <Button onClick={submit} disabled={busy || !value.trim()}>
            {busy ? "..." : confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
