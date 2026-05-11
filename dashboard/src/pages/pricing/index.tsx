import { Nav } from "../landing/ui/nav";
import { Footer } from "../landing/ui/footer";
import { Pricing as PricingTiers } from "../landing/ui/pricing";
import { ComparisonTable } from "./ui/comparison-table";

export function PricingRoute() {
  return (
    <div data-theme="eop" className="min-h-screen relative overflow-x-hidden">
      <div className="eop-bg-grid" aria-hidden />
      <div className="eop-bg-glow" aria-hidden />
      <div className="relative z-10">
        <Nav />
        <main className="pt-[68px]">
          <PricingTiers />
          <ComparisonTable />
        </main>
        <Footer />
      </div>
    </div>
  );
}
