// Безопасный markdown через react-markdown.
//
// react-markdown рендерит дерево элементов, не использует dangerouslySetInnerHTML
// и по умолчанию не пропускает HTML — поэтому XSS невозможен даже на untrusted
// LLM-выводах. Custom-классы передаём через `components` props без потери
// type-safety и без regex-парсинга.
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

const NUMERIC_RE = /^-?[\d.,]+\s*(%|сек|мин|ч|секунд|минут|часов|часа|симв|событий)?$|^-?[\d.,]+%/;
const BORDER = "hsl(var(--border))";

export function Markdown({ source }: { source: string }) {
  return (
    <div className="prose-eop">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ children }) => (
            <h1 className="mt-2 mb-4 text-2xl font-bold tracking-tight">{children}</h1>
          ),
          h2: ({ children }) => (
            <h2 className="mt-6 mb-3 text-lg font-semibold tracking-tight flex items-center gap-2">
              {children}
            </h2>
          ),
          h3: ({ children }) => (
            <h3 className="mt-5 mb-2 text-base font-semibold tracking-tight">{children}</h3>
          ),
          a: ({ href, children }) => (
            <a
              href={href}
              className="text-[hsl(var(--accent))] underline underline-offset-2 hover:opacity-90"
              target={href?.startsWith("http") ? "_blank" : undefined}
              rel={href?.startsWith("http") ? "noopener noreferrer" : undefined}
            >
              {children}
            </a>
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
          table: ({ children }) => (
            <div className="my-4 overflow-x-auto rounded-lg border" style={{ borderColor: BORDER }}>
              <table className="w-full min-w-[480px] border-collapse text-sm">{children}</table>
            </div>
          ),
          thead: ({ children }) => (
            <thead className="bg-muted/30" style={{ borderBottom: `1px solid ${BORDER}` }}>
              {children}
            </thead>
          ),
          tbody: ({ children }) => <tbody>{children}</tbody>,
          tr: ({ children }) => (
            <tr className="border-b last:border-b-0" style={{ borderColor: BORDER }}>
              {children}
            </tr>
          ),
          th: ({ children }) => (
            <th
              className="px-3 py-2 text-left font-medium text-foreground align-top"
              style={{ borderRight: `1px solid ${BORDER}` }}
            >
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td
              className="px-3 py-2 text-foreground/90 align-top"
              style={{ borderRight: `1px solid ${BORDER}` }}
            >
              {children}
            </td>
          ),
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

// react-markdown передаёт в `children` массив с разными node-типами.
// Для классификации numeric/non-numeric достаточно flatten в строку.
function childrenToString(children: React.ReactNode): string {
  if (typeof children === "string") return children;
  if (typeof children === "number") return String(children);
  if (Array.isArray(children)) return children.map(childrenToString).join("");
  return "";
}
