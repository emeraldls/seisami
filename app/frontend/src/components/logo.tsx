import React from "react";

export const Logo = ({ className }: { className?: string }) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 120 120"
      role="img"
      aria-label="Seisami logo"
      className={className}
    >
      <defs>
        <linearGradient id="bgGradient" x1="0" x2="1" y1="0" y2="1">
          <stop offset="0%" stopColor="#050505" />
          <stop offset="100%" stopColor="#1b1b1b" />
        </linearGradient>

        <radialGradient id="halo" cx="0.6" cy="0.2" r="1">
          <stop offset="0%" stopColor="#FFFFFF" stopOpacity="0.6" />
          <stop offset="75%" stopColor="#FFFFFF" stopOpacity="0.05" />
          <stop offset="100%" stopColor="#FFFFFF" stopOpacity="0" />
        </radialGradient>

        <linearGradient id="pulse" x1="0" x2="1" y1="0" y2="0">
          <stop offset="0%" stopColor="#FFFFFF" stopOpacity="0.9" />
          <stop offset="100%" stopColor="#7a7a7a" stopOpacity="0.9" />
        </linearGradient>

        <linearGradient id="signal" x1="0" x2="1" y1="0" y2="1">
          <stop offset="0%" stopColor="#FFFFFF" stopOpacity="0.95" />
          <stop offset="100%" stopColor="#c9c9c9" stopOpacity="0.8" />
        </linearGradient>

        <filter id="glow" x="-30%" y="-30%" width="160%" height="160%">
          <feGaussianBlur stdDeviation="6" result="softGlow" />
          <feMerge>
            <feMergeNode in="softGlow" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>

      <rect width="120" height="120" rx="26" fill="url(#bgGradient)" />
      <rect width="120" height="120" rx="26" fill="url(#halo)" />

      <g transform="translate(14 20)">
        <path
          d="M0 42 H92"
          stroke="rgba(255,255,255,0.12)"
          strokeWidth="2"
          strokeDasharray="6 6"
        />
        <path
          d="M8 42 C18 10, 40 10, 52 42 C64 74, 86 74, 96 42"
          stroke="url(#signal)"
          strokeWidth="12"
          strokeLinecap="round"
          fill="none"
          filter="url(#glow)"
        />
        <path
          d="M10 42 C20 24, 40 28, 52 42 C64 56, 84 60, 94 42"
          stroke="url(#pulse)"
          strokeWidth="6"
          strokeLinecap="round"
          fill="none"
        />
        <circle cx="18" cy="34" r="4" fill="#FFFFFF" opacity="0.65" />
        <circle cx="52" cy="50" r="5" fill="#FFFFFF" opacity="0.85" />
        <circle cx="86" cy="34" r="4" fill="#FFFFFF" opacity="0.65" />
      </g>
    </svg>
  );
};
