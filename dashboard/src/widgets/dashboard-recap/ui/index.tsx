import { useTranslation } from "react-i18next";

export function RecapCard() {
  const { t } = useTranslation("app");
  return (
    <div className="eop-recap">
      <div className="recap-tag">{t("dashboard.recap_tag")}</div>
      <div className="recap-text">{t("dashboard.recap_placeholder")}</div>
    </div>
  );
}
