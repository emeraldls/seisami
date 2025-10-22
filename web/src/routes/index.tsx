import { createFileRoute } from "@tanstack/react-router";
import { Header } from "../components/Header";
import { Hero } from "../components/Hero";
import { Features } from "../components/Features";
import { CTA } from "../components/CTA";
import { Footer } from "../components/Footer";

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
