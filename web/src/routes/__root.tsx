import { HeadContent, Scripts, createRootRoute } from "@tanstack/react-router";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";
import { TanStackDevtools } from "@tanstack/react-devtools";

import appCss from "../styles.css?url";

const baseUrl = "https://seisami.com";

export const Route = createRootRoute({
  head: () => ({
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
        href: `${baseUrl}/api/sitemap.xml`,
      },
    ],
  }),

  shellComponent: RootDocument,
});

function RootDocument({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <HeadContent />
      </head>
      <body className="">
        {children}
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
