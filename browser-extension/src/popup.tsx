import React from "react";
import ReactDOM from "react-dom/client";
import "@eop/ui/styles.css";
import { Button, Card, CardContent, CardHeader, CardTitle } from "@eop/ui";

function Popup() {
  return (
    <div className="w-72 p-4 space-y-3">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Eye of Providence</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <p className="text-xs text-muted-foreground">Phase 0 skeleton</p>
          <Button size="sm" className="w-full">Open dashboard</Button>
        </CardContent>
      </Card>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(<Popup />);
