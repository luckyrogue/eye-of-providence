import { useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { useQueryClient } from "@tanstack/react-query";
import { Eyebrow, Stepper } from "@eop/ui";
import { createInvite, createTeam, teamsKeys } from "@/entities/team";
import { useMutationToast } from "@/shared/hooks";
import { SESSION_KEYS } from "@/shared/lib/session-storage";
import { ONBOARDING_STEPS, type OnboardingStep } from "./model/steps";
import { CompanyStep } from "./ui/company-step";
import { EventStep } from "./ui/event-step";
import { InstallStep } from "./ui/install-step";
import { InviteStep } from "./ui/invite-step";

export function Onboarding({
  initialStep,
  initialTeamID,
  onFinish,
}: {
  initialStep: OnboardingStep;
  initialTeamID: string | null;
  onFinish: () => void;
}) {
  const { t } = useTranslation(["onboarding", "errors"]);
  const [step, setStep] = useState<OnboardingStep>(initialStep);
  const [teamID, setTeamID] = useState<string | null>(initialTeamID);
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const [inviteEmailSent, setInviteEmailSent] = useState(false);
  const [busy, setBusy] = useState(false);
  const runToast = useMutationToast();
  const qc = useQueryClient();

  const stepDefs = ONBOARDING_STEPS.map((s) => ({ key: s.id, label: t(s.i18nKey) }));

  const inviteUrl = inviteCode ? `${window.location.origin}/?invite=${inviteCode}` : "";

  async function createCompany(name: string) {
    if (!name.trim()) return;
    setBusy(true);
    const r = await runToast(createTeam(name.trim()), {
      error: t("errors:team_create_failed"),
    });
    setBusy(false);
    if (r) {
      setTeamID(r.id);
      localStorage.setItem(SESSION_KEYS.team, r.id);
      // Invalidate teams cache: AppLayout читает useTeams() для решения "редиректить
      // ли на /onboarding из-за пустых команд". Без invalidation остаётся stale empty.
      await qc.invalidateQueries({ queryKey: teamsKeys.all });
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
    if (inviteUrl) void navigator.clipboard.writeText(inviteUrl);
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
          {step === "company" && <CompanyStep busy={busy} onSubmit={createCompany} />}
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
