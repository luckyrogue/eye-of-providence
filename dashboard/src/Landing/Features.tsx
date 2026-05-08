import { Eyebrow } from "@eop/ui";
import { Activity, Brain, Code2, Lock, Sparkles, Zap } from "lucide-react";

const FEATURES = [
  {
    Icon: Brain,
    title: "AI vs manual attribution",
    desc: "Automatically detect what came from Copilot, Cursor, Claude Code, ChatGPT — and what you typed yourself.",
  },
  {
    Icon: Activity,
    title: "Real coding time",
    desc: "Active focus minutes, not just calendar time. Idle, context-switching, and reading are measured separately.",
  },
  {
    Icon: Code2,
    title: "By language & project",
    desc: "Heatmaps, breakdowns by repo and file extension. See where AI helps and where you write everything by hand.",
  },
  {
    Icon: Lock,
    title: "Privacy by design",
    desc: "We never store your prompts, code, or screen content. Only event metadata. Self-host the whole stack if you want.",
  },
  {
    Icon: Sparkles,
    title: "AI-generated reports",
    desc: "Weekly and monthly summaries from Gemini — what changed, where you sped up, what to focus on next.",
  },
  {
    Icon: Zap,
    title: "Built for teams",
    desc: "Invite-only teams, role-based access, per-project breakdowns for engineering managers.",
  },
];

export function Features() {
  return (
    <section id="features" className="py-24">
      <div className="mx-auto max-w-6xl px-6">
        <div className="max-w-2xl mb-16">
          <Eyebrow>Features</Eyebrow>
          <h2 className="display-head text-4xl md:text-5xl mt-3">
            Everything you need to <em>understand</em> your AI workflow.
          </h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {FEATURES.map(({ Icon, title, desc }) => (
            <div key={title} className="rounded-xl border bg-card p-6 card-hover">
              <div className="h-10 w-10 rounded-lg bg-secondary flex items-center justify-center mb-4">
                <Icon className="h-5 w-5" />
              </div>
              <h3 className="font-display font-bold tracking-tight text-lg">{title}</h3>
              <p className="text-sm text-muted-foreground mt-2 leading-relaxed">{desc}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
