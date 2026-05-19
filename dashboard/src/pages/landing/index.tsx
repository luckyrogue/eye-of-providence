import { PreviewProvider } from "@/shared/content";
import { MarketingPricingTiers } from "@/widgets/marketing-pricing-tiers";
import { MarketingLayout } from "@/widgets/marketing-layout";
import { Attribution } from "./ui/attribution";
import { CTASection } from "./ui/cta-section";
import { Hero } from "./ui/hero";
import { HowItWorks } from "./ui/how-it-works";
import { Integrations } from "./ui/integrations";
import { Measure } from "./ui/measure";
import { PreviewBanner } from "./ui/preview-banner";
import { Privacy } from "./ui/privacy";

export function Landing() {
  return (
    <PreviewProvider>
      <PreviewBanner />
      <MarketingLayout>
        <Hero />
        <Measure />
        <HowItWorks />
        <Attribution />
        <Privacy />
        <Integrations />
        <MarketingPricingTiers />
        <CTASection />
      </MarketingLayout>
    </PreviewProvider>
  );
}
