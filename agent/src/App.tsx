import { Button } from "@eop/ui";

export default function App() {
  return (
    <main className="min-h-screen p-6 flex flex-col gap-4">
      <h1 className="text-2xl font-semibold">Eye of Providence — agent</h1>
      <p className="text-sm text-muted-foreground">
        Phase 0 skeleton. Native watchers и ingest pipeline появятся в Phase 1.
      </p>
      <div>
        <Button>OK</Button>
      </div>
    </main>
  );
}
