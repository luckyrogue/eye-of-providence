import { Component, type ErrorInfo, type ReactNode } from "react";

type Props = { children: ReactNode };
type State = { error: Error | null };

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    // Лог в консоль; внешняя телеметрия (Sentry/Posthog) подключается позднее.
    console.error("[eop] unhandled error", error, info.componentStack);
  }

  reset = () => {
    this.setState({ error: null });
  };

  render() {
    if (this.state.error) {
      return (
        <div className="min-h-screen flex items-center justify-center p-6 bg-background">
          <div className="max-w-md w-full rounded-lg border bg-card p-6 space-y-4">
            <div>
              <h1 className="text-lg font-semibold">Что-то сломалось</h1>
              <p className="text-sm text-muted-foreground">
                Произошла ошибка интерфейса. Попробуй обновить страницу или вернуться на главную.
              </p>
            </div>
            <pre className="text-xs bg-muted/50 rounded p-2 overflow-auto max-h-40 whitespace-pre-wrap">
              {this.state.error.message}
            </pre>
            <div className="flex gap-2">
              <button
                onClick={() => location.reload()}
                className="rounded-md bg-primary text-primary-foreground px-3 py-1.5 text-sm"
              >
                Обновить
              </button>
              <button
                onClick={() => {
                  this.reset();
                  location.assign("/");
                }}
                className="rounded-md border px-3 py-1.5 text-sm"
              >
                На главную
              </button>
            </div>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
