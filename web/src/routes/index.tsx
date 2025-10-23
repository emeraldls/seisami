import { createFileRoute } from "@tanstack/react-router";
import { Header } from "../components/header";
import { Hero } from "../components/hero";
import { Features } from "../components/features";
import { CTA } from "../components/cta";
import { Footer } from "../components/footer";

export const Route = createFileRoute("/")({ component: Index });

function Index() {
  return (
    <div>
      <Header />
      <Hero />
      <Features />
      <CTA />
      <Footer />
    </div>
  );
}
