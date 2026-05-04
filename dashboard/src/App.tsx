import { Card, CardContent, CardDescription, CardHeader, CardTitle, Button } from "@eop/ui";

export default function App() {
  return (
    <main className="min-h-screen bg-background p-8">
      <div className="mx-auto max-w-5xl space-y-6">
        <header>
          <h1 className="text-3xl font-bold tracking-tight">Eye of Providence</h1>
          <p className="text-muted-foreground">Phase 0 — skeleton dashboard</p>
        </header>

        <Card>
          <CardHeader>
            <CardTitle>AI vs manual ratio</CardTitle>
            <CardDescription>Charts появятся в Phase 5</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Подключи агент, и через 24 часа здесь появится первый отчёт.
            </p>
          </CardContent>
        </Card>

        <div>
          <Button>Generate AI report (placeholder)</Button>
        </div>
      </div>
    </main>
  );
}
