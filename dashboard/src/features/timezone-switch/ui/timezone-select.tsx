import { useState } from "react";
import { useTranslation } from "react-i18next";
import { SimpleSelect } from "@eop/ui";
import { getTz, setTz, UNIQUE_TIMEZONES } from "../../../shared/lib/tz";
export function TimezoneSelect() {
  const { t } = useTranslation("common");
  const [tz, setLocalTz] = useState(getTz());
  function change(value: string) {
    setLocalTz(value);
    setTz(value);
  }
  return (
    <SimpleSelect
      value={tz}
      onValueChange={change}
      triggerClassName="w-full"
      options={[
        ...(UNIQUE_TIMEZONES.find((x) => x.value === tz)
          ? []
          : [{ value: tz, label: t("settings.timezone_current", { tz }) }]),
        ...UNIQUE_TIMEZONES.map((tzo) => ({ value: tzo.value, label: tzo.label })),
      ]}
    />
  );
}
