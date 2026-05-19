import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useLanguages } from "@/entities/event";
import { reshapeLanguages } from "../lib/reshape-languages";

export function LanguageBars() {
  const { t } = useTranslation("app");
  const lang = useLanguages(7);
  const langs = useMemo(() => reshapeLanguages(lang.data), [lang.data]);
  return (
    <div className="eop-card col-span-12 min-[1181px]:col-span-7">
      <div className="card-head">
        <div>
          <div className="card-title">{t("dashboard.langs_title")}</div>
          <div className="card-sub">{t("dashboard.langs_sub")}</div>
        </div>
      </div>
      {langs.length === 0 ? (
        <div className="text-[13px] text-muted-foreground py-4">{t("dashboard.no_data_yet")}</div>
      ) : (
        <div className="langs">
          {langs.map((l) => (
            <div key={l.name} className="lang-row">
              <span className="lang-name">{l.name}</span>
              <span className="lang-time">{l.time}</span>
              <div className="lang-stack">
                <i style={{ width: `${l.ai}%`, background: "hsl(var(--accent))" }} />
                <i style={{ width: `${l.manual}%`, background: "#4ade80", opacity: 0.65 }} />
              </div>
              <span className="lang-ratio">
                {l.ai}% {t("dashboard.lang_ratio_ai")}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
