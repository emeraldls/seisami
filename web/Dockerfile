# Build Stage
FROM node:20-alpine AS builder

WORKDIR /build

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate


# Copy package files first (for better caching)
COPY package.json pnpm-lock.yaml ./

# Install dependencies (production + dev)
RUN pnpm install --frozen-lockfile

# Copy the rest of the application
COPY . .

# Build client and server bundles
RUN NODE_OPTIONS="--max-old-space-size=4096" pnpm run build

# Production Stage
FROM node:20-alpine AS runtime

WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

COPY --from=builder /build/package.json /build/pnpm-lock.yaml ./

# Install only production dependencies
RUN pnpm install --prod --frozen-lockfile

COPY --from=builder /build/.output ./.output

EXPOSE 3000

ENV NODE_ENV=production

CMD ["node", ".output/server/index.mjs"]