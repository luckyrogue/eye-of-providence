import { useTranslation } from "react-i18next";

export function AuthOAuthSeparator() {
  const { t } = useTranslation("auth");
  return (
    <div className="relative flex items-center">
      <div className="flex-1 border-t border-border" />
      <span className="mx-2 text-xs uppercase tracking-wider text-muted-foreground">
        {t("oauth.separator_or")}
      </span>
      <div className="flex-1 border-t border-border" />
    </div>
  );
}
