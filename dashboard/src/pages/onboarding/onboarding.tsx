import { useState } from "react";
import { Eyebrow, Stepper } from "@eop/ui";
import { createInvite, createTeam } from "../../entities/team";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";
import { CompanyStep } from "./ui/company-step";
import { InstallStep } from "./ui/install-step";
import { InviteStep } from "./ui/invite-step";
import { EventStep } from "./ui/event-step";

// Порядок шагов из ROADMAP.md:
//   создай команду → установи агент → пригласи людей → первое событие.
// install идёт ПЕРЕД invite, потому что приглашённые поставят агент сами,
// а владелец должен убедиться что у него самого всё работает первым.
type Step = "company" | "install" | "invite" | "event";

const STEPS = [
  { key: "company", label: "Компания" },
  { key: "install", label: "Агент" },
  { key: "invite", label: "Команда" },
  { key: "event", label: "Событие" },
];

export function Onboarding({
  initialStep,
  initialTeamID,
  onFinish,
}: {
  initialStep: Step;
  initialTeamID: string | null;
  onFinish: () => void;
}) {
  const [step, setStep] = useState<Step>(initialStep);
  const [teamID, setTeamID] = useState<string | null>(initialTeamID);
  const [teamName, setTeamName] = useState("");
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const [inviteEmailSent, setInviteEmailSent] = useState(false);
  const [busy, setBusy] = useState(false);
  const runToast = useMutationToast();

  const inviteUrl = inviteCode ? `${window.location.origin}/?invite=${inviteCode}` : "";

  async function createCompany() {
    if (!teamName.trim()) return;
    setBusy(true);
    const r = await runToast(createTeam(teamName.trim()), {
      success: "Компания создана",
      error: "Не удалось создать компанию",
    });
    setBusy(false);
    if (r) {
      setTeamID(r.id);
      localStorage.setItem("eop_team", r.id);
      setStep("install");
    }
  }

  async function generateInvite(email?: string) {
    if (!teamID) return;
    setBusy(true);
    const r = await runToast(createInvite(teamID, email), {
      success: email ? "Приглашение отправлено" : undefined,
      error: "Не удалось создать invite",
    });
    setBusy(false);
    if (r) {
      setInviteCode(r.code);
      if (email && r.sent) setInviteEmailSent(true);
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
          <Eyebrow>Welcome aboard</Eyebrow>
          <h1 className="display-head text-4xl md:text-5xl mt-3">
            Let's set up <em>your company</em>.
          </h1>
          <p className="mt-3 text-sm text-muted-foreground">
            Четыре шага, ~5 минут. Любой шаг можно пропустить и вернуться позже.
          </p>
        </div>

        <Stepper steps={STEPS} current={step} className="max-w-md mx-auto" />

        <div className="mt-6">
          {step === "company" && (
            <CompanyStep name={teamName} setName={setTeamName} busy={busy} onSubmit={createCompany} />
          )}
          {step === "install" && (
            <InstallStep onSkip={() => setStep("invite")} onContinue={() => setStep("invite")} />
          )}
          {step === "invite" && (
            <InviteStep
              busy={busy}
              inviteUrl={inviteUrl}
              inviteEmailSent={inviteEmailSent}
              onGenerate={generateInvite}
              onCopy={copyInvite}
              onSkip={() => setStep("event")}
              onContinue={() => setStep("event")}
            />
          )}
          {step === "event" && <EventStep onFinish={onFinish} onSkip={onFinish} />}
        </div>
      </div>
    </div>
  );
}
