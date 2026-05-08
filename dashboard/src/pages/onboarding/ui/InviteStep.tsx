import { useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { ArrowRight, Copy, Loader2, Mail, SkipForward } from "lucide-react";

export function InviteStep({
  busy,
  inviteUrl,
  onGenerate,
  onCopy,
  onSkip,
  onContinue,
}: {
  busy: boolean;
  inviteUrl: string;
  onGenerate: () => void;
  onCopy: () => void;
  onSkip: () => void;
  onContinue: () => void;
}) {
  const [copied, setCopied] = useState(false);
  function copy() {
    onCopy();
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }
  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Mail className="h-5 w-5" />
          Пригласите команду
        </CardTitle>
        <CardDescription>
          Сгенерируйте invite-ссылку и отправьте её первым участникам. Без агента и команды дашборд будет пустым.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!inviteUrl ? (
          <Button onClick={onGenerate} disabled={busy} className="w-full">
            {busy ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
            Сгенерировать invite-ссылку
          </Button>
        ) : (
          <>
            <div className="flex items-center gap-2">
              <input
                readOnly
                value={inviteUrl}
                className="flex-1 rounded-md border bg-background px-3 py-2 text-xs font-mono"
              />
              <Button size="sm" variant="outline" onClick={copy}>
                <Copy className="h-3.5 w-3.5 mr-1" />
                {copied ? "Скопировано!" : "Скопировать"}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground font-mono">
              Действительна 7 дней. До 10 регистраций. Отозвать в Команда → Участники.
            </p>
          </>
        )}
        <div className="flex justify-between pt-2">
          <Button variant="ghost" onClick={onSkip} className="text-muted-foreground">
            <SkipForward className="h-3.5 w-3.5 mr-1.5" />
            Пропустить
          </Button>
          {inviteUrl && (
            <Button onClick={onContinue}>
              Дальше <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
