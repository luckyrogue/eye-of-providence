import { Component, type ErrorInfo, type ReactNode } from "react";
import i18next from "i18next";
import { Button } from "@eop/ui";
type Props = {
  children: ReactNode;
};
type State = {
  error: Error | null;
};
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };
  static getDerivedStateFromError(error: Error): State {
    return { error };
  }
  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("[eop] unhandled error", error, info.componentStack);
  }
  reset = () => {
    this.setState({ error: null });
  };
  render() {
    if (this.state.error) {
      const t = (key: string, fallback: string) =>
        i18next.exists(key) ? i18next.t(key) : fallback;
      return (
        <div className="min-h-screen flex items-center justify-center p-6 bg-background">
          <div className="max-w-md w-full rounded-lg border bg-card p-6 space-y-4">
            <div>
              <h1 className="text-lg font-semibold">
                {t("error.boundary_title", "Something went wrong")}
              </h1>
              <p className="text-sm text-muted-foreground">
                {t(
                  "error.boundary_lead",
                  "An interface error occurred. Try refreshing the page or returning home.",
                )}
              </p>
            </div>
            <pre className="text-xs bg-muted/50 rounded p-2 overflow-auto max-h-40 whitespace-pre-wrap">
              {this.state.error.message}
            </pre>
            <div className="flex gap-2">
              <Button size="sm" onClick={() => location.reload()}>
                {t("error.boundary_reload", "Reload")}
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  this.reset();
                  location.assign("/");
                }}
              >
                {t("error.boundary_home", "Home")}
              </Button>
            </div>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
