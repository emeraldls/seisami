import React from "react";
import { cn } from "@/lib/utils";

interface DitherBackgroundProps extends React.HTMLAttributes<HTMLDivElement> {
  opacity?: number;
}

export function DitherBackground({
  className,
  opacity = 0.1,
  ...props
}: DitherBackgroundProps) {
  return (
    <div
      className={cn(
        "pointer-events-none absolute inset-0 z-0 overflow-hidden",
        className
      )}
      {...props}
    >
      <div
        className="absolute inset-0"
        style={{
          backgroundImage: `radial-gradient(#000 1px, transparent 1px)`,
          backgroundSize: "4px 4px",
          opacity: opacity,
          maskImage: "linear-gradient(to bottom, black, transparent)",
          WebkitMaskImage: "linear-gradient(to bottom, black, transparent)",
        }}
      />
    </div>
  );
}
