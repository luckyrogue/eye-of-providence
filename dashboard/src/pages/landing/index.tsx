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
