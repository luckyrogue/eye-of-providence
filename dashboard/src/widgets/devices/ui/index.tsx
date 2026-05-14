import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Laptop2 } from "lucide-react";
import { useDevices } from "../../../entities/device";
import { ClaimDeviceForm } from "../../../features/device-claim";
import { DevicesTable } from "./devices-table";
export function DevicesWidget() {
  const { t } = useTranslation("developer");
  const devices = useDevices();
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Laptop2 className="h-4 w-4 text-muted-foreground" />
          <CardTitle>{t("devices_title")}</CardTitle>
        </div>
        <CardDescription>{t("devices_lead")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <DevicesTable devices={devices.data ?? []} isLoading={devices.isPending} />
        <div className="border-t pt-4">
          <div className="text-sm font-medium mb-1">{t("devices_claim_title")}</div>
          <p className="text-xs text-muted-foreground mb-3">{t("devices_claim_lead")}</p>
          <ClaimDeviceForm />
        </div>
      </CardContent>
    </Card>
  );
}
