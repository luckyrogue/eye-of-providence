import { Nav } from "./ui/nav";
import { Hero } from "./ui/hero";
import { Measure } from "./ui/measure";
import { HowItWorks } from "./ui/how-it-works";
import { Attribution } from "./ui/attribution";
import { Privacy } from "./ui/privacy";
import { Integrations } from "./ui/integrations";
import { Pricing } from "./ui/pricing";
import { CTASection } from "./ui/cta-section";
import { Footer } from "./ui/footer";

// data-theme="eop" — переключает CSS vars на warm-dark + orange палитру.
// Структура соответствует Claude artifact (Eye of Providence) 1:1: Hero →
// Measure → How → Attribution → Privacy → Integrations → Pricing → CTA → Footer.
// LogoStrip / Features / FAQ удалены — их заменили новые секции.
export function Landing() {
  return (
    <div data-theme="eop" className="min-h-screen relative overflow-x-hidden">
      <div className="eop-bg-grid" aria-hidden />
      <div className="eop-bg-glow" aria-hidden />
      <div className="eop-bg-noise" aria-hidden />
      <div className="relative z-10">
        <Nav />
        <Hero />
        <Measure />
        <HowItWorks />
        <Attribution />
        <Privacy />
        <Integrations />
        <Pricing />
        <CTASection />
        <Footer />
      </div>
    </div>
  );
}
