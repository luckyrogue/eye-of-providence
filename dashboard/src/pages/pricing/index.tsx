import { MarketingLayout } from "@/widgets/marketing-layout";
import { MarketingPricingTiers } from "@/widgets/marketing-pricing-tiers";
import { ComparisonTable } from "./ui/comparison-table";

export function PricingRoute() {
  return (
    <MarketingLayout>
      <main className="pt-[68px]">
        <MarketingPricingTiers />
        <ComparisonTable />
      </main>
    </MarketingLayout>
  );
}
