import { createFileRoute } from "@tanstack/react-router";
import { HeroSection } from "@/components/hero-section";
import { BentoGrid } from "@/components/bento-grid";
import { HowItWorks } from "@/components/how-it-works";
import { FeatureDetail } from "@/components/feature-detail";
import { FAQ } from "@/components/faq";
import { FooterSection } from "@/components/footer-section";
import { NavbarSection } from "@/components/navbar-section";

export const Route = createFileRoute("/")({ component: Index });

function Index() {
  return (
    <div className="min-h-screen bg-white dark:bg-black text-black dark:text-white selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black">
      <NavbarSection />
      <HeroSection />
      <BentoGrid />
      <HowItWorks />
      <FeatureDetail />
      <FAQ />
      <FooterSection />
    </div>
  );
}
