import { useTranslation } from "react-i18next";
import { useDevices } from "../../entities/device";
import { useTeams } from "../../entities/team";
import { useProfile } from "../../entities/user";
import { LinkedAccountsCard } from "../../features/linked-accounts";
import type { Locale } from "../../shared/i18n";
import { SettingsDangerZone } from "./ui/settings-danger-zone";
import { SettingsEmailCard } from "./ui/settings-email-card";
import { SettingsLocaleTimezoneCard } from "./ui/settings-locale-timezone-card";
import { SettingsNameCard } from "./ui/settings-name-card";
import { SettingsPageHead } from "./ui/settings-page-head";
import { SettingsPasskeyCard } from "./ui/settings-passkey-card";
import { SettingsPasswordCard } from "./ui/settings-password-card";
import { SettingsPrivacyCard } from "./ui/settings-privacy-card";
import { SettingsProfileCard } from "./ui/settings-profile-card";

export function Settings({ onWiped }: { onWiped: () => void }) {
  const { i18n } = useTranslation("common");
  const profile = useProfile();
  const teams = useTeams();
  const devices = useDevices();

  const hasPassword = profile.data?.has_password ?? false;
  const uiLocale = i18n.resolvedLanguage ?? "ru";
  const userLocale = (profile.data?.locale ?? uiLocale) as Locale;
  const firstName = profile.data?.display_name ?? "";
  const lastName = profile.data?.last_name ?? "";

  return (
    <div className="space-y-4">
      <SettingsPageHead />

      <SettingsProfileCard
        profile={profile.data}
        isProfileError={profile.isError}
        teamsLoading={teams.isLoading}
        teamsCount={teams.data?.length ?? 0}
        devicesLoading={devices.isLoading}
        devicesCount={devices.data?.length ?? 0}
        uiLocale={uiLocale}
        userLocale={userLocale}
      />

      <SettingsNameCard displayName={firstName} lastName={lastName} />

      <LinkedAccountsCard />

      <SettingsPasskeyCard />

      <SettingsEmailCard hasPassword={hasPassword} currentEmail={profile.data?.email} />

      <SettingsPasswordCard hasPassword={hasPassword} />

      <SettingsLocaleTimezoneCard />

      <SettingsPrivacyCard />

      <SettingsDangerZone onWiped={onWiped} />
    </div>
  );
}
