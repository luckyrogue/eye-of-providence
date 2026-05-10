import { useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { useQueryClient } from "@tanstack/react-query";
import { Eyebrow, Stepper } from "@eop/ui";
import { createInvite, createTeam, teamsKeys } from "../../entities/team";
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

export function Onboarding({
  initialStep,
  initialTeamID,
  onFinish,
}: {
  initialStep: Step;
  initialTeamID: string | null;
  onFinish: () => void;
}) {
  const { t } = useTranslation(["onboarding", "errors"]);
  const [step, setStep] = useState<Step>(initialStep);
  const [teamID, setTeamID] = useState<string | null>(initialTeamID);
  const [teamName, setTeamName] = useState("");
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const [inviteEmailSent, setInviteEmailSent] = useState(false);
  const [busy, setBusy] = useState(false);
  const runToast = useMutationToast();
  const qc = useQueryClient();

  const stepDefs = [
    { key: "company", label: t("onboarding:steps.company") },
    { key: "install", label: t("onboarding:steps.install") },
    { key: "invite", label: t("onboarding:steps.invite") },
    { key: "event", label: t("onboarding:steps.event") },
  ];

  const inviteUrl = inviteCode ? `${window.location.origin}/?invite=${inviteCode}` : "";

  async function createCompany() {
    if (!teamName.trim()) return;
    setBusy(true);
    const r = await runToast(createTeam(teamName.trim()), {
      error: t("errors:team_create_failed"),
    });
    setBusy(false);
    if (r) {
      setTeamID(r.id);
      localStorage.setItem("eop_team", r.id);
      // Invalidate teams cache: AppLayout читает useTeams() для решения "редиректить
      // ли на /onboarding из-за пустых команд". Без invalidation остаётся stale empty.
      qc.invalidateQueries({ queryKey: teamsKeys.list });
      qc.invalidateQueries({ queryKey: teamsKeys.beta });
      setStep("install");
    }
  }

  async function generateInvite(email?: string) {
    if (!teamID) return;
    setBusy(true);
    const r = await runToast(createInvite(teamID, email), {
      error: t("errors:invite_create_failed"),
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
          <Eyebrow>{t("onboarding:eyebrow")}</Eyebrow>
          <h1 className="display-head text-4xl md:text-5xl mt-3">
            <Trans i18nKey="onboarding:heading" components={{ em: <em /> }} />
          </h1>
          <p className="mt-3 text-sm text-muted-foreground">{t("onboarding:lead")}</p>
        </div>

        <Stepper steps={stepDefs} current={step} className="max-w-md mx-auto" />

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
