export const ONBOARDING_STEPS = [
  { id: "company", i18nKey: "onboarding:steps.company" },
  { id: "install", i18nKey: "onboarding:steps.install" },
  { id: "invite", i18nKey: "onboarding:steps.invite" },
  { id: "event", i18nKey: "onboarding:steps.event" },
] as const;
export type OnboardingStep = (typeof ONBOARDING_STEPS)[number]["id"];
