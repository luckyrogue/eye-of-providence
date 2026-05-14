import { useTranslation } from "react-i18next";
import {
  Avatar,
  AvatarFallback,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  getInitials,
} from "@eop/ui";
import {
  Calendar,
  KeyRound,
  Languages,
  Mail,
  Phone,
  ScanFace,
  ShieldCheck,
  Users as UsersIcon,
  Monitor,
  User,
} from "lucide-react";
import type { Profile } from "../../../entities/user";
import { formatShortDisplayName } from "../../../shared/lib/display-name";
import { LOCALE_LABELS, type Locale } from "../../../shared/i18n";
import { formatJoinDate } from "../lib/format-join-date";
import { GithubGlyph } from "./github-glyph";
import { ProfileStat } from "./profile-stat";
export function SettingsProfileCard({
  profile,
  isProfileError,
  teamsLoading,
  teamsCount,
  devicesLoading,
  devicesCount,
  uiLocale,
  userLocale,
}: {
  profile: Profile | undefined;
  isProfileError: boolean;
  teamsLoading: boolean;
  teamsCount: number;
  devicesLoading: boolean;
  devicesCount: number;
  uiLocale: string;
  userLocale: Locale;
}) {
  const { t } = useTranslation("common");
  const firstName = profile?.display_name ?? "";
  const lastName = profile?.last_name ?? "";
  const fullName = formatShortDisplayName(firstName, lastName, profile?.email ?? "—");
  const initials =
    firstName.trim() && lastName.trim()
      ? getInitials(`${firstName.trim()} ${lastName.trim()}`)
      : getInitials(fullName);
  const hasPassword = profile?.has_password ?? false;
  const isAdmin = profile?.global_role === "super_admin";
  const provider = profile?.provider;
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <User className="h-4 w-4 text-muted-foreground" />
          <CardTitle>{t("settings.profile_title")}</CardTitle>
        </div>
        <CardDescription>{t("settings.profile_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        {profile ? (
          <div className="space-y-5">
            <div className="flex items-center gap-4">
              <Avatar className="h-14 w-14">
                <AvatarFallback className="text-lg">{initials}</AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1">
                <div className="text-base font-medium leading-tight truncate">{fullName}</div>
                {profile.email && (
                  <div className="text-sm text-muted-foreground truncate flex items-center gap-1.5 mt-0.5">
                    <Mail className="h-3.5 w-3.5 shrink-0" />
                    {profile.email}
                  </div>
                )}
                <div className="flex flex-wrap gap-1.5 mt-2">
                  {provider && (
                    <span className="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-mono text-muted-foreground">
                      {provider === "github" ? (
                        <GithubGlyph className="h-3 w-3" />
                      ) : (
                        <ScanFace className="h-3 w-3" />
                      )}
                      {t(`settings.provider_${provider}` as never, { defaultValue: provider })}
                    </span>
                  )}
                  {isAdmin && (
                    <span
                      className="inline-flex items-center gap-1 rounded-full border border-primary/50 bg-primary/10 px-1.5 py-0.5 text-xs font-mono text-primary max-sm:gap-0 sm:px-2 sm:gap-1"
                      title={t("settings.role_super_admin")}
                    >
                      <ShieldCheck className="h-3.5 w-3.5 shrink-0 sm:h-3 sm:w-3" aria-hidden />
                      <span className="sr-only sm:not-sr-only sm:inline">
                        {t("settings.role_super_admin")}
                      </span>
                    </span>
                  )}
                </div>
              </div>
            </div>

            <dl className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {profile.created_at && (
                <ProfileStat
                  icon={<Calendar className="h-4 w-4" />}
                  label={t("settings.member_since")}
                  value={formatJoinDate(profile.created_at, uiLocale)}
                />
              )}
              <ProfileStat
                icon={<Languages className="h-4 w-4" />}
                label={t("settings.language_label")}
                value={LOCALE_LABELS[userLocale] ?? userLocale}
              />
              <ProfileStat
                icon={<UsersIcon className="h-4 w-4" />}
                label={t("settings.teams_label")}
                value={teamsLoading ? "…" : String(teamsCount)}
              />
              <ProfileStat
                icon={<Monitor className="h-4 w-4" />}
                label={t("settings.devices_label")}
                value={devicesLoading ? "…" : String(devicesCount)}
              />
              {profile.phone && (
                <ProfileStat
                  icon={<Phone className="h-4 w-4" />}
                  label={t("settings.phone_label")}
                  value={profile.phone}
                />
              )}
              {profile.github_login && (
                <ProfileStat
                  icon={<GithubGlyph className="h-4 w-4" />}
                  label="GitHub"
                  value={`@${profile.github_login}`}
                />
              )}
              <ProfileStat
                icon={<KeyRound className="h-4 w-4" />}
                label={t("settings.auth_label")}
                value={
                  hasPassword ? t("settings.auth_password_set") : t("settings.auth_oauth_only")
                }
              />
            </dl>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            {isProfileError ? t("settings.profile_load_failed") : t("settings.profile_loading")}
          </p>
        )}
      </CardContent>
    </Card>
  );
}
