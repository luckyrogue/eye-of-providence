import { useState } from "react";
import { Copy } from "lucide-react";
import { Button } from "./button";
export function SecretField({
  value,
  copyLabel,
  copiedLabel,
  onCopy,
}: {
  value: string;
  copyLabel: string;
  copiedLabel: string;
  onCopy?: () => void;
}) {
  const [copied, setCopied] = useState(false);
  function copy() {
    void navigator.clipboard.writeText(value);
    setCopied(true);
    onCopy?.();
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
        {copied ? copiedLabel : copyLabel}
      </Button>
    </div>
  );
}
