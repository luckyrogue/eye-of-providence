import { Nav } from "../landing/ui/nav";
import { Footer } from "../landing/ui/footer";
import { Pricing as PricingTiers } from "../landing/ui/pricing";
import { ComparisonTable } from "./ui/comparison-table";
export function PricingRoute() {
  return (
    <div className="min-h-screen relative overflow-x-hidden">
      <Nav />
      <main className="pt-[68px]">
        <PricingTiers />
        <ComparisonTable />
      </main>
      <Footer />
    </div>
  );
}
