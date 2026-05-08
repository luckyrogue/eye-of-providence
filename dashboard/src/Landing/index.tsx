import { Nav } from "./Nav";
import { Hero } from "./Hero";
import { LogoStrip } from "./LogoStrip";
import { Features } from "./Features";
import { HowItWorks } from "./HowItWorks";
import { Pricing } from "./Pricing";
import { FAQ } from "./FAQ";
import { CTASection } from "./CTASection";
import { Footer } from "./Footer";

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
