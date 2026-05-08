import { Nav } from "./ui/nav";
import { Hero } from "./ui/hero";
import { LogoStrip } from "./ui/logo-strip";
import { Features } from "./ui/features";
import { HowItWorks } from "./ui/how-it-works";
import { Pricing } from "./ui/pricing";
import { FAQ } from "./ui/faq";
import { CTASection } from "./ui/cta-section";
import { Footer } from "./ui/footer";

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
