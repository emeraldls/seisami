import { createFileRoute } from "@tanstack/react-router";
import { Hero } from "@/components/hero";
import { FeatureSet } from "@/components/feature-set";
import { CallToAction } from "@/components/call-to-action";
import { Footer } from "@/components/footer";
import { Navbar } from "@/components/navbar";

export const Route = createFileRoute("/")({ component: Index });

function Index() {
  return (
    <div>
      <Navbar />
      <Hero />
      <FeatureSet />
      <CallToAction />
      <Footer />
    </div>
  );
}
