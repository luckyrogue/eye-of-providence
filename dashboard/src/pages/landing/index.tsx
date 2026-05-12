import { PreviewProvider } from "../../shared/content";
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
import { PreviewBanner } from "./ui/preview-banner";

// Eop theme выставлен глобально на documentElement (см. index.html inline-script),
// bg-grid/glow — на body::before/::after (см. ui/styles.css). Этот компонент
// просто стакает секции в order, который соответствует Claude artifact 1:1.
//
// PreviewProvider exposes ?preview=<slug> to all child useContent() calls
// so the admin "Preview" button can overlay one draft slug on top of the
// otherwise-published landing page (Phase 4 CMS-lite).
export function Landing() {
  return (
    <PreviewProvider>
      <div className="min-h-screen relative overflow-x-hidden">
        <PreviewBanner />
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
    </PreviewProvider>
  );
}
