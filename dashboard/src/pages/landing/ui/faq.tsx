import { Eyebrow } from "@eop/ui";

const FAQ_ITEMS = [
  {
    q: "What does the agent actually collect?",
    a: "Only metadata: focused app, idle/active state, paste sizes, language extensions, AI provider names. Never prompts, code content, screenshots, or keystrokes.",
  },
  {
    q: "How is AI use detected?",
    a: "Multiple signals — the active app and URL (e.g. claude.ai, ChatGPT), IDE plugin events (Copilot accept, Cursor agent), and clipboard heuristics for pasted code from AI windows. All classification happens on your machine.",
  },
  {
    q: "Can I self-host?",
    a: "Yes. The backend is Go, dashboard is React, storage is Postgres + ClickHouse + Redis. Docker compose works out of the box. You can run the entire stack on a single VM.",
  },
  {
    q: "Is my data sold or shared?",
    a: "No. We don't run ads, we don't sell aggregate data, we don't train models on your events. If we ever change that, we'd ask first.",
  },
  {
    q: "What about my team's privacy?",
    a: "Team admins see aggregates per member. They cannot see individual prompts, files, or content. Members can pause tracking anytime and export or delete all their data.",
  },
  {
    q: "When does free end?",
    a: "Solo and self-hosted are free forever. Team will eventually have paid tiers — but during beta, all team features are free with no time limit.",
  },
];

export function FAQ() {
  return (
    <section id="faq" className="py-24 border-t bg-muted/20">
      <div className="mx-auto max-w-3xl px-6">
        <div className="mb-12">
          <Eyebrow>FAQ</Eyebrow>
          <h2 className="display-head text-4xl md:text-5xl mt-3">
            Quick answers, <em>before you ask</em>.
          </h2>
        </div>
        <div className="space-y-3">
          {FAQ_ITEMS.map((it) => (
            <details key={it.q} className="rounded-xl border bg-card group">
              <summary className="cursor-pointer list-none p-5 flex items-start justify-between gap-4 font-medium">
                <span>{it.q}</span>
                <span className="font-mono text-xl text-muted-foreground transition-transform group-open:rotate-45">+</span>
              </summary>
              <div className="px-5 pb-5 text-sm text-muted-foreground leading-relaxed">{it.a}</div>
            </details>
          ))}
        </div>
      </div>
    </section>
  );
}
