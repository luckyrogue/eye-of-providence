import { useEffect, useState } from "react";
import ReactDOM from "react-dom/client";
import "@eop/ui/styles.css";
import { Button, Card, CardContent, CardHeader, CardTitle } from "@eop/ui";
import { fetchDevToken, setConfig, clearConfig } from "./api";

function Popup() {
  const [token, setToken] = useState<string | undefined>();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    chrome.storage.local.get(["eop_token"]).then((r) => setToken(r.eop_token));
  }, []);

  async function login() {
    setBusy(true);
    setError(null);
    try {
      const t = await fetchDevToken();
      await setConfig(t);
      setToken(t);
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }

  async function logout() {
    await clearConfig();
    setToken(undefined);
  }

  async function flushNow() {
    setBusy(true);
    try {
      await chrome.runtime.sendMessage({ type: "flush-now" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="w-80 p-4 space-y-3">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Eye of Providence</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {error && <div className="text-xs text-destructive">{error}</div>}
          {token ? (
            <>
              <div className="text-xs text-muted-foreground">Подключено к eop-api.rysdavletov.org.</div>
              <div className="flex gap-2">
                <Button size="sm" className="flex-1" onClick={flushNow} disabled={busy}>
                  Flush now
                </Button>
                <Button size="sm" variant="outline" onClick={logout}>
                  Logout
                </Button>
              </div>
            </>
          ) : (
            <>
              <div className="text-xs text-muted-foreground">
                Dev login (без OAuth). Backend по умолчанию: <code>eop-api.rysdavletov.org</code>.
              </div>
              <Button size="sm" className="w-full" onClick={login} disabled={busy}>
                {busy ? "..." : "Get dev token"}
              </Button>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(<Popup />);
