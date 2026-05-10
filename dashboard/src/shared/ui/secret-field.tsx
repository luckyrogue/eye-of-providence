import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { Copy } from "lucide-react";

// Read-only поле с кнопкой "Copy". Используется для одноразового показа
// secret'а (API token plaintext, webhook secret) — после закрытия модалки
// значение становится недоступно.
export function SecretField({ value }: { value: string }) {
  const { t } = useTranslation("developer");
  const [copied, setCopied] = useState(false);
  function copy() {
    void navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }
  return (
    <div className="flex items-center gap-2">
      <input
        readOnly
        value={value}
        className="flex-1 rounded-md border bg-background px-3 py-2 text-xs font-mono"
      />
      <Button size="sm" variant="outline" onClick={copy}>
        <Copy className="h-3.5 w-3.5 mr-1" />
        {copied ? t("tokens_copied") : t("tokens_copy")}
      </Button>
    </div>
  );
}
