import { useTranslation } from "react-i18next";
import { Chrome, Cpu, Code2 } from "lucide-react";
import type { Device } from "../../../entities/device";

const ICON: Record<string, typeof Chrome> = {
  ext: Chrome,
  agent: Cpu,
  ide: Code2,
};

export function DeviceKindBadge({ kind }: { kind: Device["kind"] }) {
  const { t } = useTranslation("developer");
  const Icon = ICON[kind] ?? Chrome;
  const label =
    kind === "ext" || kind === "agent" || kind === "ide"
      ? t(`devices_kind_${kind}`)
      : t("devices_kind_unknown");
  return (
    <span className="inline-flex items-center gap-1.5 text-sm">
      <Icon className="h-3.5 w-3.5 text-muted-foreground" aria-hidden />
      <span>{label}</span>
    </span>
  );
}
