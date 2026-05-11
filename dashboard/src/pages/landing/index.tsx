import { Nav } from "./ui/nav";
import { Hero } from "./ui/hero";
import { LogoStrip } from "./ui/logo-strip";
import { Features } from "./ui/features";
import { HowItWorks } from "./ui/how-it-works";
import { Pricing } from "./ui/pricing";
import { FAQ } from "./ui/faq";
import { CTASection } from "./ui/cta-section";
import { Footer } from "./ui/footer";

// data-theme="eop" — переключает CSS vars на warm-dark + orange палитру
// (см. ui/src/shared/styles.css [data-theme="eop"]). Скоупится только на
// landing/pricing, не задевая dashboard-internal страницы.
export function Landing() {
  return (
    <div data-theme="eop" className="min-h-screen relative overflow-x-hidden">
      <div className="eop-bg-grid" aria-hidden />
      <div className="eop-bg-glow" aria-hidden />
      <div className="relative z-10">
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
    </div>
  );
}
