import { createContext, useCallback, useContext, useRef, useState, type ReactNode } from "react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./alert-dialog";
import { buttonVariants } from "./button";
import { cn } from "./cn";

// ConfirmDialog — модальная замена `confirm()` поверх shadcn AlertDialog.
// Поддерживает два режима:
// - simple: confirm/cancel
// - typeToConfirm: юзер должен ввести строку чтобы кнопка стала active —
//   для destructive-операций ("Введи название чтобы удалить").
//
// Используем AlertDialog (а не Dialog) — даёт role="alertdialog" для скрин-ридеров.

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

  function handleOpenChange(o: boolean) {
    if (!o) {
      setTyped("");
      onClose();
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          {description && <AlertDialogDescription>{description}</AlertDialogDescription>}
        </AlertDialogHeader>
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
        <AlertDialogFooter>
          <AlertDialogCancel disabled={busy}>{cancelText}</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleConfirm}
            disabled={busy || !canConfirm}
            className={cn(destructive && buttonVariants({ variant: "destructive" }))}
          >
            {busy ? "..." : confirmText}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

// useConfirm — императивный хук поверх ConfirmDialog для лёгкой замены `confirm()`.
// Использование:
//   const confirm = useConfirm();
//   if (await confirm({ title: "Удалить?", destructive: true })) { ... }

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
