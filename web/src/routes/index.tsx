import { createFileRoute } from "@tanstack/react-router";
import { HeroSection } from "@/components/hero-section";
import { FeatureSet } from "@/components/feature-set";
import { CallToAction } from "@/components/call-to-action";
import { Footer } from "@/components/footer";
import { Navbar } from "@/components/navbar";

export const Route = createFileRoute("/")({ component: Index });

function Index() {
  return (
    <div>
      <Navbar />
      <HeroSection />
      <FeatureSet />
      <CallToAction />
      <Footer />
    </div>
  );
}
