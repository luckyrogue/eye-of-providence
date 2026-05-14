import ReactMarkdown from "react-markdown";
const NUMERIC_RE = /^-?[\d.,]+\s*(%|сек|мин|ч|секунд|минут|часов|часа|симв|событий)?$|^-?[\d.,]+%/;
export function Markdown({ source }: { source: string }) {
  return (
    <div className="prose-eop">
      <ReactMarkdown
        components={{
          h1: ({ children }) => (
            <h1 className="mt-2 mb-4 text-2xl font-bold tracking-tight">{children}</h1>
          ),
          h2: ({ children }) => (
            <h2 className="mt-6 mb-3 text-lg font-semibold tracking-tight flex items-center gap-2">
              {children}
            </h2>
          ),
          p: ({ children }) => (
            <p className="mb-3 leading-relaxed text-foreground/90">{children}</p>
          ),
          ul: ({ children }) => <ul className="mb-4 space-y-1.5 pl-1">{children}</ul>,
          li: ({ children }) => (
            <li className="flex gap-2 text-foreground/90">
              <span className="text-purple-500/60 mt-1">•</span>
              <span className="flex-1">{children}</span>
            </li>
          ),
          code: ({ children }) => (
            <code className="rounded bg-purple-500/10 px-1.5 py-0.5 text-[0.85em] font-mono text-purple-700 dark:text-purple-300">
              {children}
            </code>
          ),
          em: ({ children }) => <em className="text-muted-foreground">{children}</em>,
          strong: ({ children }) => {
            const text = childrenToString(children);
            const isNumeric = NUMERIC_RE.test(text.trim());
            const className = isNumeric
              ? "text-foreground bg-amber-500/10 rounded px-1 py-0.5 font-semibold"
              : "font-semibold text-foreground";
            return <strong className={className}>{children}</strong>;
          },
        }}
      >
        {source}
      </ReactMarkdown>
    </div>
  );
}
function childrenToString(children: React.ReactNode): string {
  if (typeof children === "string") return children;
  if (typeof children === "number") return String(children);
  if (Array.isArray(children)) return children.map(childrenToString).join("");
  return "";
}
