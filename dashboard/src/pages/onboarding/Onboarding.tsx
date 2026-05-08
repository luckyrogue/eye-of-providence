import { useState } from "react";
import { Eyebrow, Stepper } from "@eop/ui";
import { createInvite, createTeam } from "../../entities/team";
import { useMutationToast } from "../../shared/hooks/useMutationToast";
import { CompanyStep } from "./ui/CompanyStep";
import { InviteStep } from "./ui/InviteStep";
import { InstallStep } from "./ui/InstallStep";
import { DoneStep } from "./ui/DoneStep";

type Step = "company" | "invite" | "install" | "done";

const STEPS = [
  { key: "company", label: "Компания" },
  { key: "invite", label: "Команда" },
  { key: "install", label: "Агент" },
  { key: "done", label: "Готово" },
];

export function Onboarding({ onFinish }: { onFinish: () => void }) {
  const [step, setStep] = useState<Step>("company");
  const [teamID, setTeamID] = useState<string | null>(null);
  const [teamName, setTeamName] = useState("");
  const [inviteCode, setInviteCode] = useState<string | null>(null);
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
      setStep("invite");
    }
  }

  async function generateInvite() {
    if (!teamID) return;
    setBusy(true);
    const r = await runToast(createInvite(teamID), { error: "Не удалось создать invite" });
    setBusy(false);
    if (r) setInviteCode(r.code);
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
            Four-минутная установка. Можно skip'нуть любой шаг и вернуться позже.
          </p>
        </div>

        <Stepper steps={STEPS} current={step} className="max-w-md mx-auto" />

        <div className="mt-6">
          {step === "company" && (
            <CompanyStep name={teamName} setName={setTeamName} busy={busy} onSubmit={createCompany} />
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
            <InstallStep onSkip={() => setStep("done")} onContinue={() => setStep("done")} />
          )}
          {step === "done" && <DoneStep onFinish={onFinish} />}
        </div>
      </div>
    </div>
  );
}
