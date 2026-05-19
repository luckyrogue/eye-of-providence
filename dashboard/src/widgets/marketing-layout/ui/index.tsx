import type { ReactNode } from "react";
import { Footer } from "./footer";
import { Nav } from "./nav";

export { Footer } from "./footer";
export { Nav } from "./nav";

export function MarketingLayout({
  children,
  showNav = true,
  showFooter = true,
}: {
  children: ReactNode;
  showNav?: boolean;
  showFooter?: boolean;
}) {
  return (
    <div className="min-h-screen relative overflow-x-hidden">
      {showNav && <Nav />}
      {children}
      {showFooter && <Footer />}
    </div>
  );
}
