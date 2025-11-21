import { HeadContent, Scripts, createRootRoute } from "@tanstack/react-router";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";
import { TanStackDevtools } from "@tanstack/react-devtools";

import appCss from "../styles.css?url";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const baseUrl = "https://seisami.hooklytics.com";

export const Route = createRootRoute({
  head: () => ({
    title: "Seisami - Voice-Driven Task Management",
    meta: [
      {
        charSet: "utf-8",
      },
      {
        name: "viewport",
        content: "width=device-width, initial-scale=1",
      },
      {
        name: "description",
        content:
          "Audio-powered productivity. Record, transcribe, and manage tasks with AI. Open source, local-first, no account required.",
      },
      {
        name: "keywords",
        content:
          "task management, voice recording, AI productivity, kanban board, transcription, open source, offline-first",
      },
      {
        name: "author",
        content: "Seisami Team",
      },
      {
        name: "robots",
        content: "index, follow",
      },
      {
        property: "og:title",
        content: "Seisami - Voice-Driven Task Management",
      },
      {
        property: "og:description",
        content:
          "Audio-powered productivity platform. Record tasks, transcribe automatically, and manage with AI.",
      },
      {
        property: "og:url",
        content: baseUrl,
      },
      {
        property: "og:type",
        content: "website",
      },
      {
        property: "og:image",
        content: `${baseUrl}/og-image.png`,
      },
      {
        name: "twitter:card",
        content: "summary_large_image",
      },
      {
        name: "twitter:title",
        content: "Seisami - Voice-Driven Task Management",
      },
      {
        name: "twitter:description",
        content:
          "Audio-powered productivity. Record, transcribe, and manage tasks with AI.",
      },
      {
        name: "twitter:image",
        content: `${baseUrl}/og-image.png`,
      },
      {
        name: "twitter:creator",
        content: "@seisami",
      },
      {
        httpEquiv: "x-ua-compatible",
        content: "IE=edge",
      },
      {
        name: "theme-color",
        content: "#000000",
      },
    ],
    links: [
      {
        rel: "stylesheet",
        href: appCss,
      },
      {
        rel: "canonical",
        href: baseUrl,
      },
      {
        rel: "sitemap",
        type: "application/xml",
        href: "/sitemap.xml",
      },
      {
        rel: "icon",
        type: "image/x-icon",
        href: "/favicon.ico",
      },
      {
        rel: "icon",
        type: "image/png",
        sizes: "16x16",
        href: "/favicon-16x16.png",
      },
      {
        rel: "icon",
        type: "image/png",
        sizes: "32x32",
        href: "/favicon-32x32.png",
      },
      {
        rel: "apple-touch-icon",
        sizes: "180x180",
        href: "/apple-touch-icon.png",
      },
      {
        rel: "manifest",
        href: "/site.webmanifest",
      },
    ],
  }),

  shellComponent: RootDocument,
});

function RootDocument({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient(); 
  return (
    <html lang="en">
      <head>
        <HeadContent />
      </head>
      <body className="">
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
        <TanStackDevtools
          config={{
            position: "bottom-right",
          }}
          plugins={[
            {
              name: "Tanstack Router",
              render: <TanStackRouterDevtoolsPanel />,
            },
          ]}
        />
        <Scripts />
      </body>
    </html>
  );
}
