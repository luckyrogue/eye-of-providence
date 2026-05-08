import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { ArrowRight, Eye, Rocket } from "lucide-react";

const NEXT_STEPS = [
  "Откройте дашборд — пока он будет пустой",
  "Покодьте 5-10 минут с агентом — метрики появятся в реальном времени",
  "Через 7 дней доступен AI-отчёт через Gemini",
];

export function DoneStep({ onFinish }: { onFinish: () => void }) {
  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Rocket className="h-5 w-5 text-purple-500" />
          Всё готово.
        </CardTitle>
        <CardDescription>
          Дашборд начнёт показывать метрики, как только агент пришлёт первое событие.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="space-y-2 text-sm">
          {NEXT_STEPS.map((s) => (
            <li key={s} className="flex items-center gap-2">
              <Eye className="h-4 w-4 text-muted-foreground" />
              <span>{s}</span>
            </li>
          ))}
        </ul>
        <Button onClick={onFinish} className="w-full mt-6 h-11">
          Перейти в дашборд <ArrowRight className="h-4 w-4 ml-2" />
        </Button>
      </CardContent>
    </Card>
  );
}
