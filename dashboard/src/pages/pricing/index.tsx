import { Nav } from "../landing/ui/nav";
import { Footer } from "../landing/ui/footer";
import { Pricing as PricingTiers } from "../landing/ui/pricing";
import { ComparisonTable } from "./ui/comparison-table";

export function PricingRoute() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <Nav />
      <main>
        <PricingTiers />
        <ComparisonTable />
      </main>
      <Footer />
    </div>
  );
}
