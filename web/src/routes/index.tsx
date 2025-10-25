import { createFileRoute } from "@tanstack/react-router";
import { HeroSection } from "@/components/hero-section";
import { FeatureSet } from "@/components/feature-set";
import { CallToAction } from "@/components/call-to-action";
import { FooterSection } from "@/components/footer-section";
import { NavbarSection } from "@/components/navbar-section";

export const Route = createFileRoute("/")({ component: Index });

function Index() {
  return (
    <div>
      <NavbarSection />
      <HeroSection />
      <FeatureSet />
      <CallToAction />
      <FooterSection />
    </div>
  );
}
