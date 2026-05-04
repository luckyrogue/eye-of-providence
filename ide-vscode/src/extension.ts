import * as vscode from "vscode";

export function activate(context: vscode.ExtensionContext) {
  console.log("[eop] vscode extension activated");

  // Phase 3: snapshot/diff per save, attribution events to local agent.
  const sub = vscode.workspace.onDidSaveTextDocument((doc) => {
    console.debug("[eop] saved", doc.fileName);
  });
  context.subscriptions.push(sub);
}

export function deactivate() {}
