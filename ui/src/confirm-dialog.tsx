import { useState, type ReactNode } from "react";
import { Button } from "./button";
import { Modal } from "./modal";

// ConfirmDialog — модальная замена `confirm()` и `prompt()` для confirm-by-typing.
// Поддерживает два режима:
// - simple: просто кнопка confirm/cancel
// - typeToConfirm: юзер должен ввести строку чтобы кнопка стала active —
//   для destructive-операций ("Введи название чтобы удалить").

export function ConfirmDialog({
  open,
  onClose,
  title,
  description,
  confirmText = "Подтвердить",
  cancelText = "Отмена",
  destructive,
  typeToConfirm,
  onConfirm,
  busy,
}: {
  open: boolean;
  onClose: () => void;
  title: ReactNode;
  description?: ReactNode;
  confirmText?: string;
  cancelText?: string;
  destructive?: boolean;
  typeToConfirm?: string;
  onConfirm: () => void;
  busy?: boolean;
}) {
  const [typed, setTyped] = useState("");
  const canConfirm = !typeToConfirm || typed === typeToConfirm;

  function handleConfirm() {
    if (!canConfirm) return;
    onConfirm();
    setTyped("");
  }

  function handleClose() {
    setTyped("");
    onClose();
  }

  return (
    <Modal open={open} onClose={handleClose} className="max-w-md">
      <div className="p-6 space-y-4">
        <h3 className="font-display text-xl tracking-tight">{title}</h3>
        {description && <div className="text-sm text-muted-foreground">{description}</div>}
        {typeToConfirm && (
          <div className="space-y-1">
            <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
              Введи «{typeToConfirm}» чтобы подтвердить
            </label>
            <input
              autoFocus
              value={typed}
              onChange={(e) => setTyped(e.target.value)}
              className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>
        )}
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="outline" onClick={handleClose} disabled={busy}>
            {cancelText}
          </Button>
          <Button
            variant={destructive ? "destructive" : "default"}
            onClick={handleConfirm}
            disabled={busy || !canConfirm}
          >
            {busy ? "..." : confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  );
}

// useConfirm — императивный хук поверх ConfirmDialog для лёгкой замены `confirm()`.
// Использование:
//   const confirm = useConfirm();
//   if (await confirm({ title: "Удалить?", destructive: true })) { ... }
import { createContext, useContext, useCallback, useRef } from "react";

type ConfirmFn = (opts: Omit<Parameters<typeof ConfirmDialog>[0], "open" | "onClose" | "onConfirm" | "busy">) => Promise<boolean>;

const ConfirmContext = createContext<ConfirmFn | null>(null);

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [opts, setOpts] = useState<Parameters<ConfirmFn>[0] | null>(null);
  const resolveRef = useRef<((v: boolean) => void) | null>(null);

  const confirm = useCallback<ConfirmFn>((o) => {
    return new Promise<boolean>((resolve) => {
      resolveRef.current = resolve;
      setOpts(o);
    });
  }, []);

  function close(result: boolean) {
    setOpts(null);
    resolveRef.current?.(result);
    resolveRef.current = null;
  }

  return (
    <ConfirmContext.Provider value={confirm}>
      {children}
      {opts && (
        <ConfirmDialog
          open
          onClose={() => close(false)}
          onConfirm={() => close(true)}
          {...opts}
        />
      )}
    </ConfirmContext.Provider>
  );
}

export function useConfirm(): ConfirmFn {
  const ctx = useContext(ConfirmContext);
  if (!ctx) throw new Error("useConfirm must be used inside <ConfirmProvider>");
  return ctx;
}
