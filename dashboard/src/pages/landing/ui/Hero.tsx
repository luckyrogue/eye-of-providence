import { Button, Eyebrow } from "@eop/ui";
import { ArrowRight, Check } from "lucide-react";
import { ProductPreview } from "./ProductPreview";

const TRUST_SIGNALS = ["Privacy-by-design", "Self-hostable", "No credit card"];

export function Hero() {
  return (
    <section className="relative overflow-hidden">
      <div className="dot-grid pointer-events-none absolute inset-0 [mask-image:radial-gradient(ellipse_at_center,black,transparent_70%)]" />
      <div className="absolute -top-40 left-1/2 -translate-x-1/2 h-[600px] w-[1100px] bg-gradient-to-br from-purple-500/10 via-blue-500/5 to-transparent blur-3xl pointer-events-none" />

      <div className="relative mx-auto max-w-6xl px-6 pt-20 pb-24 text-center">
        <Eyebrow className="reveal">Beta · 3 founding companies free</Eyebrow>
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
          {TRUST_SIGNALS.map((s) => (
            <span key={s} className="flex items-center gap-1.5">
              <Check className="h-3.5 w-3.5 text-foreground" /> {s}
            </span>
          ))}
        </div>

        <ProductPreview />
      </div>
    </section>
  );
}
