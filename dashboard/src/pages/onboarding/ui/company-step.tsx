import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { ArrowRight, Building2, Loader2 } from "lucide-react";

export function CompanyStep({
  name,
  setName,
  busy,
  onSubmit,
}: {
  name: string;
  setName: (v: string) => void;
  busy: boolean;
  onSubmit: () => void;
}) {
  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Building2 className="h-5 w-5" />
          Как называется ваша компания?
        </CardTitle>
        <CardDescription>
          Это будет workspace, в котором ваша команда увидит метрики. Можно изменить позже.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <input
          autoFocus
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && !busy && name.trim() && onSubmit()}
          placeholder="Acme Inc."
          maxLength={100}
          className="w-full rounded-md border bg-background px-3 py-2.5 text-base focus:outline-none focus:ring-2 focus:ring-primary"
        />
        <div className="flex justify-end">
          <Button onClick={onSubmit} disabled={busy || !name.trim()}>
            {busy ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
            Создать <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
