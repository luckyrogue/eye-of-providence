import { Eyebrow } from "@eop/ui";

const STEPS = [
  { num: "01", title: "Install the agent", desc: "One binary on macOS or Windows. VS Code & Claude Code hooks ship as separate plugins." },
  { num: "02", title: "Sign in", desc: "Email + password or GitHub OAuth. Invite-only team mode keeps your org private." },
  { num: "03", title: "Code as usual", desc: "Agent classifies events locally. Only metadata leaves your machine — never code or prompts." },
  { num: "04", title: "Read the dashboard", desc: "Live charts, language breakdown, AI-generated reports. Cancel anytime, export anytime." },
];

export function HowItWorks() {
  return (
    <section id="how" className="py-24 border-t bg-muted/20">
      <div className="mx-auto max-w-6xl px-6">
        <div className="max-w-2xl mb-16">
          <Eyebrow>How it works</Eyebrow>
          <h2 className="display-head text-4xl md:text-5xl mt-3">
            From install to insight in <em>under five minutes</em>.
          </h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {STEPS.map((s) => (
            <div key={s.num} className="rounded-xl border bg-card p-6">
              <div className="font-mono text-xs uppercase tracking-widest3 text-muted-foreground">step {s.num}</div>
              <h3 className="font-display font-bold tracking-tight text-lg mt-3">{s.title}</h3>
              <p className="text-sm text-muted-foreground mt-2 leading-relaxed">{s.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
