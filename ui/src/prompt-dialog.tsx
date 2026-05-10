import { useEffect, useState, type ReactNode } from "react";
import { Button } from "./button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./dialog";
import { Input } from "./input";

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
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-md p-6 space-y-4">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <div className="space-y-1">
          {label && (
            <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
              {label}
            </label>
          )}
          <Input
            autoFocus
            placeholder={placeholder}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && submit()}
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={busy}>
            {cancelText}
          </Button>
          <Button onClick={submit} disabled={busy || !value.trim()}>
            {busy ? "..." : confirmText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
