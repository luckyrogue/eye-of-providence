import { useTranslation } from "react-i18next";
export function SettingsPageHead() {
  const { t } = useTranslation("common");
  return (
    <div className="page-head">
      <div>
        <h1>{t("settings.title")}</h1>
        <div className="sub font-mono">{t("nav.settings")}</div>
      </div>
    </div>
  );
}
