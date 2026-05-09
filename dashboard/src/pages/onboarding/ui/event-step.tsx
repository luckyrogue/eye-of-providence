import { useEffect } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { ArrowRight, CheckCircle2, Loader2, Radio, SkipForward } from "lucide-react";
import { useOnboardingStatus } from "../../../entities/user";

// EventStep — последний шаг wizard'а: ждём первое событие из агента.
// Пока агент не пришлёт событие — крутим polling каждые 5s.
// Как только has_event === true — автоматически вызываем onFinish.
export function EventStep({ onFinish, onSkip }: { onFinish: () => void; onSkip: () => void }) {
  const status = useOnboardingStatus({ refetchInterval: 5000 });
  const hasEvent = status.data?.has_event ?? false;

  // Auto-advance: как только событие пришло, через 1.5s закрываем wizard.
  useEffect(() => {
    if (hasEvent) {
      const t = setTimeout(onFinish, 1500);
      return () => clearTimeout(t);
    }
  }, [hasEvent, onFinish]);

  if (hasEvent) {
    return (
      <Card className="card-hover reveal">
        <CardHeader>
          <CardTitle className="font-display tracking-tight flex items-center gap-2">
            <CheckCircle2 className="h-5 w-5 text-green-500" />
            Событие получено!
          </CardTitle>
          <CardDescription>Агент работает. Открываю дашборд…</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Radio className="h-5 w-5 text-primary animate-pulse" />
          Жду первое событие…
        </CardTitle>
        <CardDescription>
          Откройте редактор и попишите минуту-две. Как только агент отправит событие, дашборд оживёт.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-3 rounded-md border bg-muted/30 p-3">
          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          <span className="text-sm text-muted-foreground font-mono">polling /v1/me/onboarding-status каждые 5 сек</span>
        </div>
        <div className="rounded-md border bg-card p-3 text-xs space-y-2">
          <p className="font-medium">Не идут события? Проверь:</p>
          <ul className="list-disc pl-4 space-y-1 text-muted-foreground">
            <li>Tauri-агент запущен (иконка в menu-bar / system-tray)</li>
            <li>VS Code extension активирован — открой любой <code className="font-mono">.ts</code> файл</li>
            <li>Browser extension включён — переключись на ChatGPT / Claude</li>
          </ul>
        </div>
        <div className="flex justify-end pt-2">
          <Button variant="ghost" onClick={onSkip} className="text-muted-foreground">
            <SkipForward className="h-3.5 w-3.5 mr-1.5" />
            Открыть пустой дашборд
            <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
