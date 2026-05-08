import { Nav } from "./ui/Nav";
import { Hero } from "./ui/Hero";
import { LogoStrip } from "./ui/LogoStrip";
import { Features } from "./ui/Features";
import { HowItWorks } from "./ui/HowItWorks";
import { Pricing } from "./ui/Pricing";
import { FAQ } from "./ui/FAQ";
import { CTASection } from "./ui/CTASection";
import { Footer } from "./ui/Footer";

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
