import React from "react";
import ReactDOM from "react-dom/client";
import "@eop/ui/styles.css";
import App from "./App";
import { Landing } from "./Landing";
import { ErrorBoundary } from "./ErrorBoundary";

// Простой роутинг по pathname без react-router — для лендинга и дашборда хватает.
// `/` и `/landing` показывают marketing-страницу, всё остальное идёт в App
// (где Auth-форма сама разруливает залогиненного / нет).
function Root() {
  const path = window.location.pathname;
  if (path === "/" || path === "/landing") {
    return <Landing />;
  }
  return <App />;
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <Root />
    </ErrorBoundary>
  </React.StrictMode>
);
