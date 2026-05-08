import { useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { ArrowRight, Building2, Check, Copy, Download, Eye, Loader2, Mail, Rocket, SkipForward } from "lucide-react";
import { createTeam, createInvite } from "./api";

type Step = "company" | "invite" | "install" | "done";

export function Onboarding({ onFinish }: { onFinish: () => void }) {
  const [step, setStep] = useState<Step>("company");
  const [teamID, setTeamID] = useState<string | null>(null);
  const [teamName, setTeamName] = useState("");
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const inviteUrl = inviteCode ? `${window.location.origin}/?invite=${inviteCode}` : "";

  async function createCompany() {
    if (!teamName.trim()) return;
    setBusy(true);
    setError(null);
    try {
      const r = await createTeam(teamName.trim());
      setTeamID(r.id);
      localStorage.setItem("eop_team", r.id);
      setStep("invite");
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  async function generateInvite() {
    if (!teamID) return;
    setBusy(true);
    try {
      const r = await createInvite(teamID);
      setInviteCode(r.code);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  function copyInvite() {
    if (inviteUrl) navigator.clipboard.writeText(inviteUrl);
  }

  return (
    <div className="relative min-h-[calc(100vh-100px)]">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[420px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />

      <div className="mx-auto max-w-2xl pt-10 pb-6">
        <div className="text-center mb-10 reveal">
          <span className="eyebrow">Welcome aboard</span>
          <h1 className="display-head text-4xl md:text-5xl mt-3">
            Let's set up <em>your company</em>.
          </h1>
          <p className="mt-3 text-sm text-muted-foreground">
            Four-минутная установка. Можно skip'нуть любой шаг и вернуться позже.
          </p>
        </div>

        <Stepper current={step} />

        {error && (
          <div className="mt-4 rounded-md border border-destructive bg-destructive/10 p-3 text-sm text-destructive">
            {error}
          </div>
        )}

        <div className="mt-6">
          {step === "company" && (
            <CompanyStep
              name={teamName}
              setName={setTeamName}
              busy={busy}
              onSubmit={createCompany}
            />
          )}
          {step === "invite" && (
            <InviteStep
              busy={busy}
              inviteUrl={inviteUrl}
              onGenerate={generateInvite}
              onCopy={copyInvite}
              onSkip={() => setStep("install")}
              onContinue={() => setStep("install")}
            />
          )}
          {step === "install" && (
            <InstallStep
              onSkip={() => setStep("done")}
              onContinue={() => setStep("done")}
            />
          )}
          {step === "done" && <DoneStep onFinish={onFinish} />}
        </div>
      </div>
    </div>
  );
}

function Stepper({ current }: { current: Step }) {
  const steps: { key: Step; label: string }[] = [
    { key: "company", label: "Компания" },
    { key: "invite", label: "Команда" },
    { key: "install", label: "Агент" },
    { key: "done", label: "Готово" },
  ];
  const index = steps.findIndex((s) => s.key === current);
  return (
    <div className="flex items-center gap-2 max-w-md mx-auto">
      {steps.map((s, i) => {
        const isActive = i === index;
        const isDone = i < index;
        return (
          <div key={s.key} className="flex-1 flex items-center gap-2">
            <div
              className={`flex items-center justify-center h-7 w-7 rounded-full text-[11px] font-mono transition-colors ${
                isDone
                  ? "bg-primary text-primary-foreground"
                  : isActive
                    ? "border-2 border-primary text-primary"
                    : "border border-border text-muted-foreground"
              }`}
            >
              {isDone ? <Check className="h-3.5 w-3.5" /> : i + 1}
            </div>
            <span
              className={`text-xs font-mono uppercase tracking-widest2 ${
                isActive || isDone ? "text-foreground" : "text-muted-foreground"
              }`}
            >
              {s.label}
            </span>
            {i < steps.length - 1 && <div className="flex-1 h-px bg-border" />}
          </div>
        );
      })}
    </div>
  );
}

function CompanyStep({
  name,
  setName,
  busy,
  onSubmit,
}: {
  name: string;
  setName: (v: string) => void;
  busy: boolean;
  onSubmit: () => void;
}) {
  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Building2 className="h-5 w-5" />
          Как называется ваша компания?
        </CardTitle>
        <CardDescription>
          Это будет workspace, в котором ваша команда увидит метрики. Можно изменить позже.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <input
          autoFocus
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && !busy && name.trim() && onSubmit()}
          placeholder="Acme Inc."
          maxLength={100}
          className="w-full rounded-md border bg-background px-3 py-2.5 text-base focus:outline-none focus:ring-2 focus:ring-primary"
        />
        <div className="flex justify-end">
          <Button onClick={onSubmit} disabled={busy || !name.trim()}>
            {busy ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
            Создать <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function InviteStep({
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

function InstallStep({ onSkip, onContinue }: { onSkip: () => void; onContinue: () => void }) {
  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Download className="h-5 w-5" />
          Установите агент
        </CardTitle>
        <CardDescription>
          Без агента события не будут собираться. Можно установить позже — установка займёт ~2 минуты.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
          <a
            href="/downloads/eop-mac.dmg"
            target="_blank"
            rel="noreferrer"
            className="rounded-lg border bg-card p-4 text-center hover:bg-secondary card-hover"
          >
            <div className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">macOS</div>
            <div className="font-display font-bold mt-1">.dmg</div>
            <div className="text-[11px] text-muted-foreground mt-1">10.15+</div>
          </a>
          <a
            href="/downloads/eop-win.msi"
            target="_blank"
            rel="noreferrer"
            className="rounded-lg border bg-card p-4 text-center hover:bg-secondary card-hover"
          >
            <div className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">Windows</div>
            <div className="font-display font-bold mt-1">.msi</div>
            <div className="text-[11px] text-muted-foreground mt-1">10/11</div>
          </a>
          <a
            href="https://marketplace.visualstudio.com"
            target="_blank"
            rel="noreferrer"
            className="rounded-lg border bg-card p-4 text-center hover:bg-secondary card-hover"
          >
            <div className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">VS Code</div>
            <div className="font-display font-bold mt-1">extension</div>
            <div className="text-[11px] text-muted-foreground mt-1">marketplace</div>
          </a>
        </div>
        <div className="rounded-md border bg-muted/30 p-3 text-xs">
          <p className="text-muted-foreground">
            Также есть browser extension и git-hooks — детали в{" "}
            <a href="/docs/install" className="underline hover:text-foreground">документации</a>.
          </p>
        </div>
        <div className="flex justify-between pt-2">
          <Button variant="ghost" onClick={onSkip} className="text-muted-foreground">
            <SkipForward className="h-3.5 w-3.5 mr-1.5" />
            Установлю позже
          </Button>
          <Button onClick={onContinue}>
            Готово <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function DoneStep({ onFinish }: { onFinish: () => void }) {
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
          <li className="flex items-center gap-2">
            <Eye className="h-4 w-4 text-muted-foreground" />
            <span>Откройте дашборд — пока он будет пустой</span>
          </li>
          <li className="flex items-center gap-2">
            <Eye className="h-4 w-4 text-muted-foreground" />
            <span>Покодьте 5-10 минут с агентом — метрики появятся в реальном времени</span>
          </li>
          <li className="flex items-center gap-2">
            <Eye className="h-4 w-4 text-muted-foreground" />
            <span>Через 7 дней доступен AI-отчёт через Gemini</span>
          </li>
        </ul>
        <Button onClick={onFinish} className="w-full mt-6 h-11">
          Перейти в дашборд <ArrowRight className="h-4 w-4 ml-2" />
        </Button>
      </CardContent>
    </Card>
  );
}
