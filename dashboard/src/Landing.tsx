import { Button } from "@eop/ui";
import {
  Activity,
  ArrowRight,
  Brain,
  Check,
  Code2,
  Eye,
  FileText,
  Github,
  Lock,
  Sparkles,
  Zap,
} from "lucide-react";

export function Landing() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <Nav />
      <Hero />
      <LogoStrip />
      <Features />
      <HowItWorks />
      <Pricing />
      <FAQ />
      <CTASection />
      <Footer />
    </div>
  );
}

function Nav() {
  const isAuthed = typeof window !== "undefined" && !!localStorage.getItem("eop_user_id");
  return (
    <header className="sticky top-0 z-40 border-b header-blur">
      <div className="mx-auto max-w-6xl px-6 h-16 flex items-center justify-between">
        <a href="/" className="flex items-center gap-2.5 group">
          <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center transition-transform duration-300 ease-out-expo group-hover:rotate-[8deg]">
            <Eye className="h-4 w-4 text-primary-foreground" />
          </div>
          <span className="font-display font-bold tracking-tightest text-lg">Eye of Providence</span>
        </a>
        <nav className="hidden md:flex items-center gap-8 text-sm">
          <a href="#features" className="text-muted-foreground hover:text-foreground transition-colors">Features</a>
          <a href="#how" className="text-muted-foreground hover:text-foreground transition-colors">How it works</a>
          <a href="#pricing" className="text-muted-foreground hover:text-foreground transition-colors">Pricing</a>
          <a href="#faq" className="text-muted-foreground hover:text-foreground transition-colors">FAQ</a>
        </nav>
        <div className="flex items-center gap-2">
          {!isAuthed && (
            <a
              href="/dashboard"
              className="hidden sm:inline-flex items-center text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              Sign in
            </a>
          )}
          <Button asChild size="sm">
            <a href="/dashboard">{isAuthed ? "Open dashboard" : "Get started"}</a>
          </Button>
        </div>
      </div>
    </header>
  );
}

function Hero() {
  return (
    <section className="relative overflow-hidden">
      <div className="dot-grid pointer-events-none absolute inset-0 [mask-image:radial-gradient(ellipse_at_center,black,transparent_70%)]" />
      <div className="absolute -top-40 left-1/2 -translate-x-1/2 h-[600px] w-[1100px] bg-gradient-to-br from-purple-500/10 via-blue-500/5 to-transparent blur-3xl pointer-events-none" />

      <div className="relative mx-auto max-w-6xl px-6 pt-20 pb-24 text-center">
        <span className="eyebrow reveal">Now in beta · free</span>
        <h1 className="display-head text-5xl sm:text-6xl md:text-7xl mt-5 max-w-3xl mx-auto reveal reveal-delay-1">
          See how much you code <em>vs.</em> the AI.
        </h1>
        <p className="mt-6 text-lg text-muted-foreground max-w-xl mx-auto reveal reveal-delay-2">
          Privacy-first analytics for engineers. Track manual coding, AI assists, and where your time really goes — across IDE, browser, and CLI.
        </p>
        <div className="mt-8 flex items-center justify-center gap-3 reveal reveal-delay-3">
          <Button asChild size="lg" className="h-11 px-6">
            <a href="/dashboard" className="flex items-center gap-2">
              Get started — free
              <ArrowRight className="h-4 w-4" />
            </a>
          </Button>
          <Button asChild size="lg" variant="outline" className="h-11 px-6">
            <a href="#how">See how it works</a>
          </Button>
        </div>
        <div className="mt-12 flex items-center justify-center gap-6 text-xs text-muted-foreground font-mono reveal reveal-delay-4">
          <span className="flex items-center gap-1.5"><Check className="h-3.5 w-3.5 text-foreground" /> Privacy-by-design</span>
          <span className="flex items-center gap-1.5"><Check className="h-3.5 w-3.5 text-foreground" /> Self-hostable</span>
          <span className="flex items-center gap-1.5"><Check className="h-3.5 w-3.5 text-foreground" /> No credit card</span>
        </div>

        <ProductPreview />
      </div>
    </section>
  );
}

function ProductPreview() {
  return (
    <div className="relative mt-16 mx-auto max-w-5xl reveal reveal-delay-5">
      <div className="absolute -inset-x-8 -inset-y-4 bg-gradient-to-b from-transparent via-purple-500/5 to-blue-500/10 blur-2xl pointer-events-none" />
      <div className="relative rounded-xl border bg-card/80 backdrop-blur-sm shadow-2xl overflow-hidden">
        <div className="border-b px-4 py-2 flex items-center gap-2 bg-muted/30">
          <span className="h-2.5 w-2.5 rounded-full bg-red-500/70" />
          <span className="h-2.5 w-2.5 rounded-full bg-amber-500/70" />
          <span className="h-2.5 w-2.5 rounded-full bg-emerald-500/70" />
          <span className="ml-3 font-mono text-xs text-muted-foreground">eop.rysdavletov.org/dashboard</span>
        </div>
        <div className="p-6 space-y-4">
          <div className="grid grid-cols-3 gap-3">
            <PreviewStat label="AI ratio" value="42%" accent="purple" />
            <PreviewStat label="Active time" value="6h 18m" accent="blue" />
            <PreviewStat label="Manual" value="58%" accent="amber" />
          </div>
          <div className="rounded-lg border bg-card p-4 h-44 relative overflow-hidden">
            <span className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">Last 30 days</span>
            <FakeChart />
          </div>
        </div>
      </div>
    </div>
  );
}

function PreviewStat({ label, value, accent }: { label: string; value: string; accent: "purple" | "blue" | "amber" }) {
  const accents = {
    purple: "from-purple-500/20",
    blue: "from-blue-500/20",
    amber: "from-amber-500/20",
  };
  return (
    <div className="rounded-lg border bg-card p-3 relative overflow-hidden">
      <div className={`absolute right-0 top-0 h-16 w-16 rounded-bl-full bg-gradient-to-bl ${accents[accent]} to-transparent`} />
      <div className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">{label}</div>
      <div className="font-display text-3xl font-bold tracking-tightest tabular-nums mt-1">{value}</div>
    </div>
  );
}

function FakeChart() {
  // Cosmetic SVG чисто для preview-карточки
  const points = [10, 22, 18, 30, 26, 38, 34, 48, 42, 56, 50, 62, 58, 70, 66, 76];
  const pointsAi = [4, 8, 6, 10, 12, 14, 16, 22, 26, 30, 34, 36, 40, 44, 48, 52];
  const max = 80;
  const path = (arr: number[]) =>
    arr
      .map((v, i) => `${(i / (arr.length - 1)) * 100},${100 - (v / max) * 90}`)
      .join(" ");
  return (
    <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="absolute inset-0 w-full h-full mt-6 px-3 pb-3">
      <defs>
        <linearGradient id="manualG" x1="0" x2="0" y1="0" y2="1">
          <stop offset="0%" stopColor="hsl(220 80% 55%)" stopOpacity="0.3" />
          <stop offset="100%" stopColor="hsl(220 80% 55%)" stopOpacity="0" />
        </linearGradient>
        <linearGradient id="aiG" x1="0" x2="0" y1="0" y2="1">
          <stop offset="0%" stopColor="hsl(280 70% 60%)" stopOpacity="0.3" />
          <stop offset="100%" stopColor="hsl(280 70% 60%)" stopOpacity="0" />
        </linearGradient>
      </defs>
      <polyline points={path(points)} fill="none" stroke="hsl(220 80% 55%)" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
      <polyline points={path(pointsAi)} fill="none" stroke="hsl(280 70% 60%)" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
      <polygon points={`0,100 ${path(points)} 100,100`} fill="url(#manualG)" />
      <polygon points={`0,100 ${path(pointsAi)} 100,100`} fill="url(#aiG)" />
    </svg>
  );
}

function LogoStrip() {
  const sources = [
    { name: "VS Code", icon: <Code2 className="h-4 w-4" /> },
    { name: "Cursor", icon: <Sparkles className="h-4 w-4" /> },
    { name: "Claude Code", icon: <Brain className="h-4 w-4" /> },
    { name: "GitHub Copilot", icon: <Github className="h-4 w-4" /> },
    { name: "Browser", icon: <Activity className="h-4 w-4" /> },
    { name: "CLI", icon: <FileText className="h-4 w-4" /> },
  ];
  return (
    <section className="border-y bg-muted/20">
      <div className="mx-auto max-w-6xl px-6 py-10">
        <p className="text-center text-xs font-mono uppercase tracking-widest3 text-muted-foreground mb-6">
          Tracks every channel you actually use
        </p>
        <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-4 text-sm text-muted-foreground">
          {sources.map((s) => (
            <span key={s.name} className="flex items-center gap-2 hover:text-foreground transition-colors">
              {s.icon}
              <span className="font-medium">{s.name}</span>
            </span>
          ))}
        </div>
      </div>
    </section>
  );
}

function Features() {
  const items = [
    {
      icon: <Brain className="h-5 w-5" />,
      title: "AI vs manual attribution",
      desc: "Automatically detect what came from Copilot, Cursor, Claude Code, ChatGPT — and what you typed yourself.",
    },
    {
      icon: <Activity className="h-5 w-5" />,
      title: "Real coding time",
      desc: "Active focus minutes, not just calendar time. Idle, context-switching, and reading are measured separately.",
    },
    {
      icon: <Code2 className="h-5 w-5" />,
      title: "By language & project",
      desc: "Heatmaps, breakdowns by repo and file extension. See where AI helps and where you write everything by hand.",
    },
    {
      icon: <Lock className="h-5 w-5" />,
      title: "Privacy by design",
      desc: "We never store your prompts, code, or screen content. Only event metadata. Self-host the whole stack if you want.",
    },
    {
      icon: <Sparkles className="h-5 w-5" />,
      title: "AI-generated reports",
      desc: "Weekly and monthly summaries from Gemini — what changed, where you sped up, what to focus on next.",
    },
    {
      icon: <Zap className="h-5 w-5" />,
      title: "Built for teams",
      desc: "Invite-only teams, role-based access, per-project breakdowns for engineering managers.",
    },
  ];
  return (
    <section id="features" className="py-24">
      <div className="mx-auto max-w-6xl px-6">
        <div className="max-w-2xl mb-16">
          <span className="eyebrow">Features</span>
          <h2 className="display-head text-4xl md:text-5xl mt-3">
            Everything you need to <em>understand</em> your AI workflow.
          </h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {items.map((it) => (
            <div key={it.title} className="rounded-xl border bg-card p-6 card-hover">
              <div className="h-10 w-10 rounded-lg bg-secondary flex items-center justify-center mb-4">
                {it.icon}
              </div>
              <h3 className="font-display font-bold tracking-tight text-lg">{it.title}</h3>
              <p className="text-sm text-muted-foreground mt-2 leading-relaxed">{it.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function HowItWorks() {
  const steps = [
    {
      num: "01",
      title: "Install the agent",
      desc: "One binary on macOS or Windows. VS Code & Claude Code hooks ship as separate plugins.",
    },
    {
      num: "02",
      title: "Sign in",
      desc: "Email + password or GitHub OAuth. Invite-only team mode keeps your org private.",
    },
    {
      num: "03",
      title: "Code as usual",
      desc: "Agent classifies events locally. Only metadata leaves your machine — never code or prompts.",
    },
    {
      num: "04",
      title: "Read the dashboard",
      desc: "Live charts, language breakdown, AI-generated reports. Cancel anytime, export anytime.",
    },
  ];
  return (
    <section id="how" className="py-24 border-t bg-muted/20">
      <div className="mx-auto max-w-6xl px-6">
        <div className="max-w-2xl mb-16">
          <span className="eyebrow">How it works</span>
          <h2 className="display-head text-4xl md:text-5xl mt-3">
            From install to insight in <em>under five minutes</em>.
          </h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {steps.map((s) => (
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

function Pricing() {
  return (
    <section id="pricing" className="py-24">
      <div className="mx-auto max-w-6xl px-6">
        <div className="text-center mb-12">
          <span className="eyebrow">Pricing</span>
          <h2 className="display-head text-4xl md:text-5xl mt-3 max-w-2xl mx-auto">
            Free during beta. <em>Forever for solo devs.</em>
          </h2>
          <p className="text-muted-foreground mt-4 max-w-lg mx-auto">
            We're in early access. No credit card, no trial countdown. Team plans land later — you'll be the first to know.
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 max-w-5xl mx-auto">
          <PriceCard
            name="Solo"
            price="Free"
            period="forever"
            highlight={false}
            features={[
              "Personal dashboard",
              "Up to 3 projects",
              "30-day event history",
              "Weekly AI report",
              "Community support",
            ]}
            cta="Get started"
          />
          <PriceCard
            name="Team"
            price="$0"
            period="/seat / month — beta"
            highlight={true}
            features={[
              "Everything in Solo",
              "Unlimited projects",
              "Team analytics & roles",
              "Invite-only signups",
              "18-month event history",
              "Priority email support",
            ]}
            cta="Join the beta"
          />
          <PriceCard
            name="Self-hosted"
            price="Free"
            period="open-core"
            highlight={false}
            features={[
              "Run the whole stack",
              "Your data on your infra",
              "Docker / docker-compose",
              "Postgres + ClickHouse",
              "Community support",
            ]}
            cta="Read the docs"
          />
        </div>
      </div>
    </section>
  );
}

function PriceCard({
  name,
  price,
  period,
  features,
  highlight,
  cta,
}: {
  name: string;
  price: string;
  period: string;
  features: string[];
  highlight: boolean;
  cta: string;
}) {
  return (
    <div
      className={`relative rounded-xl border p-7 card-hover ${highlight ? "border-foreground bg-card shadow-lg" : "bg-card"}`}
    >
      {highlight && (
        <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-foreground text-background px-3 py-1 text-[10px] font-mono uppercase tracking-widest2">
          Most popular
        </span>
      )}
      <h3 className="font-display font-bold text-2xl tracking-tight">{name}</h3>
      <div className="mt-4 flex items-baseline gap-1">
        <span className="font-display text-5xl font-bold tracking-tightest tabular-nums">{price}</span>
        <span className="text-sm text-muted-foreground ml-2">{period}</span>
      </div>
      <ul className="mt-6 space-y-2.5 text-sm">
        {features.map((f) => (
          <li key={f} className="flex items-start gap-2">
            <Check className="h-4 w-4 mt-0.5 text-foreground shrink-0" />
            <span>{f}</span>
          </li>
        ))}
      </ul>
      <Button asChild className="w-full mt-7" variant={highlight ? "default" : "outline"}>
        <a href="/dashboard">{cta}</a>
      </Button>
    </div>
  );
}

function FAQ() {
  const items = [
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
  return (
    <section id="faq" className="py-24 border-t bg-muted/20">
      <div className="mx-auto max-w-3xl px-6">
        <div className="mb-12">
          <span className="eyebrow">FAQ</span>
          <h2 className="display-head text-4xl md:text-5xl mt-3">
            Quick answers, <em>before you ask</em>.
          </h2>
        </div>
        <div className="space-y-3">
          {items.map((it) => (
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

function CTASection() {
  return (
    <section className="py-24">
      <div className="mx-auto max-w-3xl px-6 text-center">
        <h2 className="display-head text-5xl md:text-6xl">
          Ship more. <em>Know how.</em>
        </h2>
        <p className="mt-5 text-muted-foreground max-w-md mx-auto">
          Free during beta. Two minutes to try. Cancel any time.
        </p>
        <Button asChild size="lg" className="mt-8 h-12 px-8">
          <a href="/dashboard" className="flex items-center gap-2">
            Get started — free
            <ArrowRight className="h-4 w-4" />
          </a>
        </Button>
      </div>
    </section>
  );
}

function Footer() {
  return (
    <footer className="border-t">
      <div className="mx-auto max-w-6xl px-6 py-12 grid grid-cols-2 md:grid-cols-4 gap-8 text-sm">
        <div className="col-span-2">
          <div className="flex items-center gap-2.5">
            <div className="h-7 w-7 rounded-md bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
              <Eye className="h-3.5 w-3.5 text-primary-foreground" />
            </div>
            <span className="font-display font-bold tracking-tightest">Eye of Providence</span>
          </div>
          <p className="text-muted-foreground mt-3 max-w-xs leading-relaxed text-xs">
            Privacy-first AI workflow analytics for engineers. Self-hostable.
          </p>
        </div>
        <div>
          <div className="font-mono text-[11px] uppercase tracking-widest2 text-muted-foreground mb-3">Product</div>
          <ul className="space-y-2 text-muted-foreground">
            <li><a href="#features" className="hover:text-foreground transition-colors">Features</a></li>
            <li><a href="#pricing" className="hover:text-foreground transition-colors">Pricing</a></li>
            <li><a href="#faq" className="hover:text-foreground transition-colors">FAQ</a></li>
            <li><a href="/dashboard" className="hover:text-foreground transition-colors">Dashboard</a></li>
          </ul>
        </div>
        <div>
          <div className="font-mono text-[11px] uppercase tracking-widest2 text-muted-foreground mb-3">Legal</div>
          <ul className="space-y-2 text-muted-foreground">
            <li><a href="/privacy" className="hover:text-foreground transition-colors">Privacy</a></li>
            <li><a href="/terms" className="hover:text-foreground transition-colors">Terms</a></li>
            <li><a href="/security" className="hover:text-foreground transition-colors">Security</a></li>
          </ul>
        </div>
      </div>
      <div className="border-t">
        <div className="mx-auto max-w-6xl px-6 py-5 flex items-center justify-between text-xs text-muted-foreground font-mono">
          <span>© {new Date().getFullYear()} Eye of Providence</span>
          <span className="flex items-center gap-1.5">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
            All systems operational
          </span>
        </div>
      </div>
    </footer>
  );
}
