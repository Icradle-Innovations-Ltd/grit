package scaffold

import (
	"fmt"
	"path/filepath"
)

func writeWebFiles(root string, opts Options) error {
	webRoot := filepath.Join(root, "apps", "web")

	files := map[string]string{
		filepath.Join(webRoot, "package.json"):             webPackageJSON(opts),
		filepath.Join(webRoot, "next.config.ts"):           webNextConfig(),
		filepath.Join(webRoot, "tailwind.config.ts"):       webTailwindConfig(),
		filepath.Join(webRoot, "postcss.config.js"):        webPostCSSConfig(),
		filepath.Join(webRoot, "tsconfig.json"):            webTSConfig(),
		filepath.Join(webRoot, "app", "globals.css"):       webGlobalCSS(),
		filepath.Join(webRoot, "app", "layout.tsx"):        webRootLayout(opts),
		filepath.Join(webRoot, "app", "page.tsx"):          webLandingPage(opts),
		filepath.Join(webRoot, "app", "error.tsx"):         webErrorPage(),
		filepath.Join(webRoot, "app", "not-found.tsx"):     webNotFoundPage(),
		filepath.Join(webRoot, "app", "global-error.tsx"):  webGlobalErrorPage(),
		filepath.Join(webRoot, "lib", "utils.ts"):          webUtils(),
		filepath.Join(webRoot, "components", "navbar.tsx"): webNavbar(opts),
		filepath.Join(webRoot, "components", "footer.tsx"): webFooter(opts),
		// v3.31.49 — DevLinks renders every URL the `grit new` welcome
		// banner prints (API, GORM Studio, Sentinel, Admin, MinIO,
		// Mailhog, ...) on the landing page, dev-only.
		filepath.Join(webRoot, "components", "dev-links.tsx"): webDevLinks(),
		filepath.Join(webRoot, "components", "providers.tsx"): webProviders(),
		// v3.31.48 -- AppChrome ships in the base scaffold (handles
		// /forms/<token> chromeless rendering for public form-share).
		// UserMenu, web-session marker, auth pages, useAuth, auth
		// shells, and the auth-aware navbar are opt-in via
		// `grit add web-auth` -- see webAuthFiles() in web_auth.go.
		filepath.Join(webRoot, "components", "AppChrome.tsx"):       webAppChrome(),
		filepath.Join(webRoot, "lib", "api.ts"):                     webAPIClient(),
		filepath.Join(webRoot, "hooks", "use-blogs.ts"):             webUseBlogsHook(),
		filepath.Join(webRoot, "app", "blog", "page.tsx"):           webBlogListPage(),
		filepath.Join(webRoot, "app", "blog", "[slug]", "page.tsx"): webBlogDetailPage(),
		// v3.31.20: public form-share page (Phase 2)
		filepath.Join(webRoot, "app", "forms", "[token]", "page.tsx"): webPublicFormPage(),
		filepath.Join(webRoot, "public", ".gitkeep"):                  "",
	}

	for path, content := range files {
		if err := writeFile(path, content); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}

	if err := writeBrandLogo(filepath.Join(webRoot, "public"), "grit_logo.png"); err != nil {
		return err
	}

	return nil
}

func webPackageJSON(opts Options) string {
	return fmt.Sprintf(`{
  "name": "@%s/web",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "rm -rf .next && next dev --port 3000",
    "build": "next build",
    "start": "next start",
    "lint": "next lint",
    "format": "prettier --write .",
    "test": "vitest run",
    "test:watch": "vitest",
    "test:ui": "vitest --ui"
  },
  "dependencies": {
    "@repo/shared": "workspace:*",
    "@tanstack/react-query": "^5.17.0",
    "axios": "^1.6.0",
    "class-variance-authority": "^0.7.0",
    "clsx": "^2.1.0",
    "lucide-react": "^0.303.0",
    "next": "^16.1.6",
    "react": "19.2.7",
    "react-dom": "19.2.7",
    "react-hook-form": "^7.49.0",
    "@hookform/resolvers": "^3.3.0",
    "js-cookie": "^3.0.5",
    "tailwind-merge": "^2.2.0",
    "tailwindcss-animate": "^1.0.7"
  },
  "devDependencies": {
    "@testing-library/jest-dom": "^6.4.0",
    "@testing-library/react": "^16.0.0",
    "@testing-library/user-event": "^14.5.0",
    "@vitejs/plugin-react": "^4.3.0",
    "@types/node": "^20.0.0",
    "@types/js-cookie": "^3.0.6",
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "autoprefixer": "^10.4.0",
    "jsdom": "^25.0.0",
    "postcss": "^8.4.0",
    "prettier": "^3.3.0",
    "prettier-plugin-tailwindcss": "^0.6.0",
    "tailwindcss": "^3.4.0",
    "typescript": "^5.3.0",
    "vitest": "^2.0.0"
  }
}
`, opts.ProjectName)
}

func webNextConfig() string {
	return `import type { NextConfig } from "next";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

// Hoist the monorepo's root .env into process.env. Next.js auto-loads
// .env only from the package's own directory, so without this the THEME
// and SOCIAL_AUTH_ENABLED values set at the root are invisible to the
// web app. Shell env wins — we only fill in unset keys.
const rootEnv = resolve(process.cwd(), "..", "..", ".env");
if (existsSync(rootEnv)) {
  for (const line of readFileSync(rootEnv, "utf8").split(/\r?\n/)) {
    const m = line.match(/^([A-Z_][A-Z0-9_]*)=(.*)$/i);
    if (!m) continue;
    if (process.env[m[1]] === undefined) process.env[m[1]] = m[2].trim();
  }
}

const nextConfig: NextConfig = {
  output: "standalone",
  reactStrictMode: true,
  // packages/shared ships TypeScript source rather than a built bundle,
  // so Next needs to run it through SWC. Otherwise imports of
  // @repo/shared/types fail with "Cannot find module" at build time.
  transpilePackages: ["@repo/shared"],
  // Mirror THEME + SOCIAL_AUTH_ENABLED from .env into the NEXT_PUBLIC_*
  // namespace so the active theme is visible to server components and
  // the client bundle. Defaults keep new apps booting without env edits.
  env: {
    NEXT_PUBLIC_THEME: process.env.THEME || "atlas",
    NEXT_PUBLIC_SOCIAL_AUTH_ENABLED: process.env.SOCIAL_AUTH_ENABLED || "true",
  },
};

export default nextConfig;
`
}

func webTailwindConfig() string {
	return `import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./lib/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "var(--bg-primary)",
        "bg-secondary": "var(--bg-secondary)",
        "bg-tertiary": "var(--bg-tertiary)",
        "bg-elevated": "var(--bg-elevated)",
        "bg-hover": "var(--bg-hover)",
        border: "var(--border)",
        foreground: "var(--text-primary)",
        "text-secondary": "var(--text-secondary)",
        "text-muted": "var(--text-muted)",
        accent: {
          DEFAULT: "var(--accent)",
          hover: "var(--accent-hover)",
        },
        success: "var(--success)",
        danger: "var(--danger)",
        warning: "var(--warning)",
        info: "var(--info)",
      },
      fontFamily: {
        // v3.28.1: --font-display + --font-mono are set per theme in the
        // root layout (Inter for atlas, Geist for aurora, Onest for pulse).
        sans: ["var(--font-display)", "system-ui", "sans-serif"],
        mono: ["var(--font-mono)", "ui-monospace", "monospace"],
        serif: ["var(--font-serif)", "Georgia", "serif"],
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
};

export default config;
`
}

func webPostCSSConfig() string {
	return `module.exports = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
};
`
}

func webTSConfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2017",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "paths": {
      "@/*": ["./*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules", "vitest.config.ts", "vitest.setup.ts", "playwright.config.ts", "**/__tests__/**", "**/*.test.ts", "**/*.test.tsx", "**/*.spec.ts", "**/*.spec.tsx", "e2e"]
}
`
}

func webGlobalCSS() string {
	return `@tailwind base;
@tailwind components;
@tailwind utilities;

/* v3.28.1 — theme-aware CSS variables. Web mirrors admin's variable
 * system so a single THEME=<name> in .env paints both surfaces. */

/* atlas (default) */
:root,
[data-theme="atlas"] {
  --bg-primary: #ffffff;
  --bg-secondary: #f8fafc;
  --bg-tertiary: #f1f5f9;
  --bg-elevated: #ffffff;
  --bg-hover: #f1f5f9;
  --border: #e2e8f0;
  --text-primary: #0f172a;
  --text-secondary: #475569;
  --text-muted: #94a3b8;
  --accent: #2563eb;
  --accent-hover: #1d4ed8;
  --success: #10b981;
  --danger: #ef4444;
  --warning: #f59e0b;
  --info: #0ea5e9;
}

[data-theme="aurora"] {
  --bg-primary: #fafaf9;
  --bg-secondary: #ffffff;
  --bg-tertiary: #f5f5f4;
  --bg-elevated: #ffffff;
  --bg-hover: #f5f5f4;
  --border: #e7e5e4;
  --text-primary: #1c1917;
  --text-secondary: #57534e;
  --text-muted: #a8a29e;
  --accent: #7c3aed;
  --accent-hover: #6d28d9;
  --success: #10b981;
  --danger: #ef4444;
  --warning: #f59e0b;
  --info: #0ea5e9;
}

[data-theme="pulse"] {
  --bg-primary: #fafaf9;
  --bg-secondary: #ffffff;
  --bg-tertiary: #f5f5f4;
  --bg-elevated: #ffffff;
  --bg-hover: #f5f5f4;
  --border: #e5e5e5;
  --text-primary: #0f0f0f;
  --text-secondary: #525252;
  --text-muted: #a3a3a3;
  --accent: #0f0f0f;
  --accent-hover: #1f1f1f;
  --success: #16a34a;
  --danger: #dc2626;
  --warning: #fbbf24;
  --info: #0284c7;
}

[data-theme="midnight"] {
  --bg-primary: #0a0a0f;
  --bg-secondary: #111118;
  --bg-tertiary: #1a1a24;
  --bg-elevated: #22222e;
  --bg-hover: #2a2a38;
  --border: #2a2a3a;
  --text-primary: #e8e8f0;
  --text-secondary: #9090a8;
  --text-muted: #606078;
  --accent: #6c5ce7;
  --accent-hover: #7c6cf7;
  --success: #00b894;
  --danger: #ff6b6b;
  --warning: #fdcb6e;
  --info: #74b9ff;
}

body {
  background-color: var(--bg-primary);
  color: var(--text-primary);
  font-family: var(--font-display), system-ui, sans-serif;
}

* {
  border-color: var(--border);
}

::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: var(--bg-primary);
}

::-webkit-scrollbar-thumb {
  background: var(--bg-hover);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

/* Blog prose styles for rendered HTML content */
.prose-blog {
  color: var(--text-primary);
  font-size: 1.0625rem;
  line-height: 1.8;
}

.prose-blog h1 {
  font-size: 2rem;
  font-weight: 700;
  margin-top: 2.5rem;
  margin-bottom: 1rem;
  letter-spacing: -0.025em;
  color: var(--text-primary);
}

.prose-blog h2 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-top: 2rem;
  margin-bottom: 0.75rem;
  letter-spacing: -0.025em;
  color: var(--text-primary);
}

.prose-blog h3 {
  font-size: 1.25rem;
  font-weight: 600;
  margin-top: 1.75rem;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.prose-blog h4 {
  font-size: 1.125rem;
  font-weight: 600;
  margin-top: 1.5rem;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.prose-blog p {
  margin-bottom: 1.25rem;
  color: var(--text-secondary);
}

.prose-blog a {
  color: var(--accent);
  text-decoration: underline;
  text-underline-offset: 2px;
  transition: color 0.15s ease;
}

.prose-blog a:hover {
  color: var(--accent-hover);
}

.prose-blog strong {
  font-weight: 600;
  color: var(--text-primary);
}

.prose-blog em {
  font-style: italic;
}

.prose-blog ul {
  list-style: disc;
  padding-left: 1.5rem;
  margin-bottom: 1.25rem;
}

.prose-blog ol {
  list-style: decimal;
  padding-left: 1.5rem;
  margin-bottom: 1.25rem;
}

.prose-blog li {
  margin-bottom: 0.375rem;
  color: var(--text-secondary);
}

.prose-blog li::marker {
  color: var(--text-muted);
}

.prose-blog blockquote {
  border-left: 3px solid var(--accent);
  padding-left: 1rem;
  margin: 1.5rem 0;
  font-style: italic;
  color: var(--text-secondary);
  background: var(--bg-secondary);
  padding: 1rem 1.25rem;
  border-radius: 0 0.5rem 0.5rem 0;
}

.prose-blog pre {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 0.75rem;
  padding: 1.25rem;
  overflow-x: auto;
  margin: 1.5rem 0;
  font-family: var(--font-jetbrains-mono), ui-monospace, monospace;
  font-size: 0.875rem;
  line-height: 1.7;
}

.prose-blog code {
  font-family: var(--font-jetbrains-mono), ui-monospace, monospace;
  font-size: 0.875em;
  background: var(--bg-elevated);
  padding: 0.15rem 0.4rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  color: var(--accent);
}

.prose-blog pre code {
  background: transparent;
  padding: 0;
  border: none;
  border-radius: 0;
  font-size: inherit;
  color: var(--text-primary);
}

.prose-blog img {
  max-width: 100%;
  height: auto;
  border-radius: 0.75rem;
  margin: 1.5rem 0;
  border: 1px solid var(--border);
}

.prose-blog hr {
  border: none;
  border-top: 1px solid var(--border);
  margin: 2rem 0;
}

.prose-blog table {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5rem 0;
}

.prose-blog th {
  text-align: left;
  padding: 0.75rem 1rem;
  border-bottom: 2px solid var(--border);
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.875rem;
}

.prose-blog td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border);
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.prose-blog tr:last-child td {
  border-bottom: none;
}
`
}

func webRootLayout(opts Options) string {
	// v3.28.1: per-theme font loading + data-theme attribute. Mirrors the
	// admin layout — see the comment there for the trade-off (build-time
	// font import means switching themes without re-scaffolding loses the
	// matching family; fine for now since runtime theme switching is rare).
	var fontImport, fontVars string
	switch opts.Theme {
	case "aurora":
		fontImport = `import { Geist, Geist_Mono } from "next/font/google";

const geist = Geist({ subsets: ["latin"], variable: "--font-display", weight: ["400", "500", "600", "700"] });
const geistMono = Geist_Mono({ subsets: ["latin"], variable: "--font-mono", weight: ["400", "500", "600"] });`
		fontVars = "${geist.variable} ${geistMono.variable}"
	case "pulse":
		fontImport = `import { Onest, DM_Serif_Display, JetBrains_Mono } from "next/font/google";

const onest = Onest({ subsets: ["latin"], variable: "--font-display", weight: ["400", "500", "600", "700"] });
const dmSerif = DM_Serif_Display({ subsets: ["latin"], variable: "--font-serif", weight: ["400"] });
const jetbrainsMono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono", weight: ["400", "500", "600"] });`
		fontVars = "${onest.variable} ${dmSerif.variable} ${jetbrainsMono.variable}"
	default: // atlas
		fontImport = `import { Inter, JetBrains_Mono } from "next/font/google";

const inter = Inter({ subsets: ["latin"], variable: "--font-display", weight: ["400", "500", "600", "700"] });
const jetbrainsMono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono", weight: ["400", "500", "600"] });`
		fontVars = "${inter.variable} ${jetbrainsMono.variable}"
	}

	return fmt.Sprintf(`import type { Metadata } from "next";
%s
import { Providers } from "@/components/providers";
import { AppChrome } from "@/components/AppChrome";
import "./globals.css";

export const metadata: Metadata = {
  title: "%s — Go + React. Built with Grit.",
  description: "A full-stack framework that combines Go backend with Next.js frontend. Build fast, ship faster.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const dataTheme = process.env.NEXT_PUBLIC_THEME || "atlas";

  return (
    <html lang="en" data-theme={dataTheme} suppressHydrationWarning>
      <body className={`+"`%s "+`font-sans antialiased`+"`"+`}>
        <Providers>
          <AppChrome>{children}</AppChrome>
        </Providers>
      </body>
    </html>
  );
}`, fontImport, opts.ProjectName, fontVars)
}

func webLandingPage(_ Options) string {
	return `"use client";

import { ArrowRight } from "lucide-react";
import Link from "next/link";
import { usePublicBlogs } from "@/hooks/use-blogs";
import { DevLinks } from "@/components/dev-links";

const DOCS_URL = "https://grit-vert.vercel.app/docs";

export default function HomePage() {
  const { data, isLoading } = usePublicBlogs(1, 3);
  const blogs = data?.blogs || [];

  return (
    <div className="flex flex-col">
      {/* Hero */}
      <section className="flex items-center justify-center">
        <div className="mx-auto max-w-2xl px-6 py-24 text-center">
          <div className="mb-8 inline-flex items-center gap-2 rounded-full border border-border bg-bg-secondary px-4 py-1.5 text-sm text-text-secondary">
            <span className="h-2 w-2 rounded-full bg-success animate-pulse" />
            <span>Go + React Full-Stack Framework</span>
          </div>

          <h1 className="text-5xl font-bold tracking-tight sm:text-6xl lg:text-7xl">
            <span className="text-accent">Grit</span>
          </h1>

          <p className="mt-6 text-lg text-text-secondary leading-relaxed max-w-lg mx-auto">
            The full-stack meta-framework that fuses Go, React, and a
            Filament-like admin panel. Scaffold entire projects, generate
            resources, and ship fast.
          </p>

          <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
            <a
              href={` + "`" + `${DOCS_URL}/getting-started/quick-start` + "`" + `}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 rounded-lg bg-accent px-6 py-3 text-sm font-semibold text-white hover:bg-accent-hover transition-colors"
            >
              Get Started <ArrowRight className="h-4 w-4" />
            </a>
            <a
              href={` + "`" + `${DOCS_URL}` + "`" + `}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 rounded-lg border border-border bg-bg-secondary px-6 py-3 text-sm font-semibold text-foreground hover:bg-bg-hover transition-colors"
            >
              Read the Docs
            </a>
          </div>

          {/* Terminal snippet */}
          <div className="mt-16 mx-auto max-w-md rounded-xl border border-border bg-bg-secondary shadow-2xl overflow-hidden text-left">
            <div className="flex items-center gap-2 border-b border-border px-4 py-2.5">
              <div className="h-2.5 w-2.5 rounded-full bg-danger/60" />
              <div className="h-2.5 w-2.5 rounded-full bg-warning/60" />
              <div className="h-2.5 w-2.5 rounded-full bg-success/60" />
              <span className="ml-2 text-[11px] text-text-muted font-mono">terminal</span>
            </div>
            <div className="p-5 font-mono text-sm space-y-1.5">
              <p><span className="text-success select-none">$ </span><span className="text-foreground">grit new my-saas</span></p>
              <p><span className="text-success select-none">$ </span><span className="text-foreground">cd my-saas && docker compose up -d</span></p>
              <p><span className="text-success select-none">$ </span><span className="text-foreground">pnpm dev</span></p>
              <p className="text-success pt-1">Ready on http://localhost:3000</p>
            </div>
          </div>
        </div>
      </section>

      {/* Recent Posts */}
      <section className="border-t border-border/50 bg-bg-secondary/30">
        <div className="mx-auto max-w-5xl px-6 py-20">
          <div className="flex items-center justify-between mb-10">
            <div>
              <h2 className="text-2xl font-bold tracking-tight">Recent Posts</h2>
              <p className="mt-1 text-sm text-text-secondary">Latest articles and updates</p>
            </div>
            <Link
              href="/blog"
              className="text-sm text-accent hover:text-accent-hover transition-colors font-medium flex items-center gap-1"
            >
              View all <ArrowRight className="h-3.5 w-3.5" />
            </Link>
          </div>

          {isLoading ? (
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="rounded-xl border border-border bg-bg-elevated overflow-hidden animate-pulse">
                  <div className="h-48 bg-bg-hover" />
                  <div className="p-5 space-y-3">
                    <div className="h-4 bg-bg-hover rounded w-3/4" />
                    <div className="h-3 bg-bg-hover rounded w-full" />
                    <div className="h-3 bg-bg-hover rounded w-2/3" />
                  </div>
                </div>
              ))}
            </div>
          ) : blogs.length > 0 ? (
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {blogs.map((blog) => (
                <Link
                  key={blog.id}
                  href={` + "`" + `/blog/${blog.slug}` + "`" + `}
                  className="group rounded-xl border border-border bg-bg-elevated overflow-hidden hover:border-accent/40 hover:shadow-lg hover:shadow-accent/5 transition-all duration-300"
                >
                  <div className="h-48 bg-bg-hover overflow-hidden">
                    {blog.image ? (
                      <img
                        src={blog.image}
                        alt={blog.title}
                        className="h-full w-full object-cover group-hover:scale-105 transition-transform duration-500"
                      />
                    ) : (
                      <div className="h-full w-full flex items-center justify-center bg-gradient-to-br from-accent/10 to-accent/5">
                        <span className="text-4xl font-bold text-accent/20">{blog.title.charAt(0)}</span>
                      </div>
                    )}
                  </div>
                  <div className="p-5">
                    <p className="text-xs text-text-muted mb-2">
                      {new Date(blog.published_at || blog.created_at).toLocaleDateString("en-US", {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                      })}
                    </p>
                    <h3 className="font-semibold text-foreground group-hover:text-accent transition-colors line-clamp-2">
                      {blog.title}
                    </h3>
                    {blog.excerpt && (
                      <p className="mt-2 text-sm text-text-secondary line-clamp-2">{blog.excerpt}</p>
                    )}
                  </div>
                </Link>
              ))}
            </div>
          ) : (
            <div className="text-center py-12">
              <p className="text-text-muted text-sm">No blog posts yet. Create your first post in the admin panel.</p>
            </div>
          )}
        </div>
      </section>

      {/* v3.31.49 -- DevLinks renders in development only. Surfaces
          every URL the ` + "`grit new`" + ` welcome banner prints (API, GORM
          Studio, Sentinel, Pulse, Admin, MinIO, Mailhog, ...) so
          the operator doesn't have to keep the terminal around. */}
      <DevLinks />
    </div>
  );
}
`
}

func webUtils() string {
	return `import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
`
}

func webNavbar(opts Options) string {
	// v3.31.48 -- this is the BASE scaffold navbar, with no auth UI.
	// Web ships without auth by default; operators add it via
	// ` + "`grit add web-auth`" + `, which overwrites this file with the
	// auth-aware version below (webNavbarWithAuth).
	//
	// v3.31.49 -- the Admin CTA is always rendered (auth or no auth)
	// so operators can bounce to /admin from the marketing site.
	return `"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu, X, Github, Shield } from "lucide-react";

const DOCS_URL = "https://grit-vert.vercel.app/docs";
const ADMIN_URL = process.env.NEXT_PUBLIC_ADMIN_URL || "http://localhost:3001";

const navLinks = [
  { href: "/", label: "Home" },
  { href: "/blog", label: "Blog" },
];

export function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const pathname = usePathname();

  return (
    <nav className="sticky top-0 z-50 border-b border-border/50 bg-background/80 backdrop-blur-lg">
      <div className="mx-auto flex h-16 max-w-5xl items-center justify-between px-6">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent/15 border border-accent/20">
            <span className="text-accent font-mono font-bold text-sm">G</span>
          </div>
          <span className="text-lg font-bold tracking-tight">` + opts.ProjectName + `</span>
        </Link>

        {/* Desktop nav */}
        <div className="hidden md:flex items-center gap-6">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className={` + "`" + `text-sm transition-colors ${
                pathname === link.href
                  ? "text-foreground font-medium"
                  : "text-text-secondary hover:text-foreground"
              }` + "`" + `}
            >
              {link.label}
            </Link>
          ))}
          <a
            href={DOCS_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-text-secondary hover:text-foreground transition-colors"
          >
            Docs
          </a>
          <a
            href="https://github.com/MUKE-coder/grit"
            target="_blank"
            rel="noopener noreferrer"
            className="text-text-secondary hover:text-foreground transition-colors"
          >
            <Github className="h-5 w-5" />
          </a>
          {/* v3.31.49 -- Admin CTA. Operators land on the marketing
              site and shouldn't have to type the admin URL by hand;
              the admin app itself gates everything behind auth. */}
          <a
            href={ADMIN_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-bg-tertiary px-3 py-1.5 text-sm font-medium text-text-secondary hover:bg-bg-hover hover:text-foreground transition-colors"
          >
            <Shield className="h-3.5 w-3.5" />
            Admin
          </a>
        </div>

        {/* Mobile hamburger */}
        <button
          onClick={() => setMobileOpen(!mobileOpen)}
          className="md:hidden p-2 text-text-secondary hover:text-foreground transition-colors"
          aria-label="Toggle menu"
        >
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {/* Mobile menu */}
      {mobileOpen && (
        <div className="md:hidden border-t border-border/50 bg-background/95 backdrop-blur-lg">
          <div className="mx-auto max-w-5xl px-6 py-4 flex flex-col gap-3">
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                onClick={() => setMobileOpen(false)}
                className={` + "`" + `text-sm py-2 transition-colors ${
                  pathname === link.href
                    ? "text-foreground font-medium"
                    : "text-text-secondary hover:text-foreground"
                }` + "`" + `}
              >
                {link.label}
              </Link>
            ))}
            <a
              href={DOCS_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm py-2 text-text-secondary hover:text-foreground transition-colors"
            >
              Docs
            </a>
            <a
              href="https://github.com/MUKE-coder/grit"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm py-2 text-text-secondary hover:text-foreground transition-colors"
            >
              GitHub
            </a>
            <a
              href={ADMIN_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-bg-tertiary px-3 py-2 text-sm font-medium text-text-secondary hover:bg-bg-hover hover:text-foreground transition-colors"
            >
              <Shield className="h-3.5 w-3.5" />
              Admin
            </a>
          </div>
        </div>
      )}
    </nav>
  );
}
`
}

// v3.31.48 -- webNavbarWithAuth is the auth-aware navbar written by
// `grit add web-auth`. It imports UserMenu (Login/Sign up CTAs when
// signed out, avatar dropdown when signed in) and replaces the
// base scaffold's plain navbar.
func webNavbarWithAuth(opts Options) string {
	return `"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu, X, Github, Shield } from "lucide-react";
import { UserMenu } from "@/components/UserMenu";

const DOCS_URL = "https://grit-vert.vercel.app/docs";
const ADMIN_URL = process.env.NEXT_PUBLIC_ADMIN_URL || "http://localhost:3001";

const navLinks = [
  { href: "/", label: "Home" },
  { href: "/blog", label: "Blog" },
];

export function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const pathname = usePathname();

  return (
    <nav className="sticky top-0 z-50 border-b border-border/50 bg-background/80 backdrop-blur-lg">
      <div className="mx-auto flex h-16 max-w-5xl items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent/15 border border-accent/20">
            <span className="text-accent font-mono font-bold text-sm">G</span>
          </div>
          <span className="text-lg font-bold tracking-tight">` + opts.ProjectName + `</span>
        </Link>

        <div className="hidden md:flex items-center gap-6">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className={` + "`" + `text-sm transition-colors ${
                pathname === link.href
                  ? "text-foreground font-medium"
                  : "text-text-secondary hover:text-foreground"
              }` + "`" + `}
            >
              {link.label}
            </Link>
          ))}
          <a
            href={DOCS_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-text-secondary hover:text-foreground transition-colors"
          >
            Docs
          </a>
          <a
            href="https://github.com/MUKE-coder/grit"
            target="_blank"
            rel="noopener noreferrer"
            className="text-text-secondary hover:text-foreground transition-colors"
          >
            <Github className="h-5 w-5" />
          </a>
          {/* v3.31.49 -- Admin CTA (always visible, even with auth). */}
          <a
            href={ADMIN_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-bg-tertiary px-3 py-1.5 text-sm font-medium text-text-secondary hover:bg-bg-hover hover:text-foreground transition-colors"
          >
            <Shield className="h-3.5 w-3.5" />
            Admin
          </a>
          <UserMenu />
        </div>

        <button
          onClick={() => setMobileOpen(!mobileOpen)}
          className="md:hidden p-2 text-text-secondary hover:text-foreground transition-colors"
          aria-label="Toggle menu"
        >
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {mobileOpen && (
        <div className="md:hidden border-t border-border/50 bg-background/95 backdrop-blur-lg">
          <div className="mx-auto max-w-5xl px-6 py-4 flex flex-col gap-3">
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                onClick={() => setMobileOpen(false)}
                className={` + "`" + `text-sm py-2 transition-colors ${
                  pathname === link.href
                    ? "text-foreground font-medium"
                    : "text-text-secondary hover:text-foreground"
                }` + "`" + `}
              >
                {link.label}
              </Link>
            ))}
            <a
              href={DOCS_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm py-2 text-text-secondary hover:text-foreground transition-colors"
            >
              Docs
            </a>
            <a
              href="https://github.com/MUKE-coder/grit"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm py-2 text-text-secondary hover:text-foreground transition-colors"
            >
              GitHub
            </a>
            <a
              href={ADMIN_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-bg-tertiary px-3 py-2 text-sm font-medium text-text-secondary hover:bg-bg-hover hover:text-foreground transition-colors"
            >
              <Shield className="h-3.5 w-3.5" />
              Admin
            </a>
            <div className="mt-2 border-t border-border/50 pt-3">
              <UserMenu />
            </div>
          </div>
        </div>
      )}
    </nav>
  );
}
`
}

func webFooter(opts Options) string {
	return `import Link from "next/link";

const DOCS_URL = "https://grit-vert.vercel.app/docs";

export function Footer() {
  return (
    <footer className="border-t border-border/50 bg-background">
      <div className="mx-auto max-w-5xl px-6 py-8">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2 text-sm text-text-muted">
            <span className="font-semibold text-text-secondary">` + opts.ProjectName + `</span>
            <span className="text-border">·</span>
            <span>Built with Grit</span>
          </div>
          <div className="flex items-center gap-6 text-sm text-text-muted">
            <a
              href="https://github.com/MUKE-coder/grit"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-foreground transition-colors"
            >
              GitHub
            </a>
            <a
              href={DOCS_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-foreground transition-colors"
            >
              Documentation
            </a>
          </div>
        </div>
        <div className="mt-6 pt-6 border-t border-border/30 text-center">
          <p className="text-xs text-text-muted">
            &copy; {new Date().getFullYear()} ` + opts.ProjectName + `. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  );
}
`
}

// v3.31.49 -- webDevLinks emits the DevLinks component that renders
// every URL printed by the `grit new` welcome banner (API, GORM
// Studio, Sentinel, Pulse, Admin, MinIO, Mailhog, ...) on the
// landing page. Wrapped in a NODE_ENV !== "production" check so
// production deploys never expose the internal port map.
func webDevLinks() string {
	return `"use client";

// v3.31.49 -- DevLinks renders a grid of the local URLs printed by
// ` + "`grit new`" + ` (API, GORM Studio, Sentinel, Admin, MinIO, Mailhog, ...)
// directly on the marketing site, so operators don't have to keep the
// terminal output around to find them.
//
// Only renders in development (NODE_ENV !== "production") so production
// marketing pages never leak the internal port map. The check happens
// at module level so the section disappears entirely from the prod
// bundle -- not just hidden behind a class.

import {
  Server,
  Database,
  Shield,
  Activity,
  HardDrive,
  Mail,
  LayoutDashboard,
  FileText,
  ExternalLink,
} from "lucide-react";

const ADMIN_URL = process.env.NEXT_PUBLIC_ADMIN_URL || "http://localhost:3001";
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface DevLink {
  title: string;
  description: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  group: "app" | "api" | "data" | "ops";
}

const LINKS: DevLink[] = [
  { title: "Admin panel",   description: "Resource CRUD, system hub, dashboard customisation",  href: ADMIN_URL,             icon: LayoutDashboard, group: "app"  },
  { title: "API root",      description: "Health check + every JSON endpoint",                   href: API_URL,               icon: Server,          group: "api"  },
  { title: "API docs",      description: "OpenAPI / Swagger UI",                                 href: API_URL + "/docs",     icon: FileText,        group: "api"  },
  { title: "GORM Studio",   description: "Visual database browser (your tables, no SQL)",        href: API_URL + "/studio",   icon: Database,        group: "data" },
  { title: "Sentinel",      description: "Security + rate-limit dashboard",                      href: API_URL + "/sentinel/ui", icon: Shield,       group: "ops"  },
  { title: "Pulse",         description: "Observability: traces, slow queries, SLO timelines",   href: API_URL + "/pulse/ui",    icon: Activity,     group: "ops"  },
  { title: "MinIO console", description: "Object storage browser (buckets, uploads)",            href: "http://localhost:9003",  icon: HardDrive,    group: "data" },
  { title: "Mailhog",       description: "Email catcher (dev only; intercepts every outbound)",  href: "http://localhost:8025",  icon: Mail,         group: "ops"  },
];

const groupAccent: Record<DevLink["group"], string> = {
  app:  "text-accent bg-accent/10",
  api:  "text-info bg-info/10",
  data: "text-success bg-success/10",
  ops:  "text-warning bg-warning/10",
};

export function DevLinks() {
  if (process.env.NODE_ENV === "production") return null;
  return (
    <section className="py-16 border-t border-border/50">
      <div className="mx-auto max-w-5xl px-6">
        <div className="mb-8">
          <p className="text-xs font-semibold uppercase tracking-wider text-text-muted">
            Local development
          </p>
          <h2 className="mt-1 text-2xl font-bold tracking-tight text-foreground">
            Developer links
          </h2>
          <p className="mt-1 text-sm text-text-secondary">
            All the dashboards and consoles your project ships with — wired to your local ports. Hidden in production.
          </p>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {LINKS.map((l) => {
            const Icon = l.icon;
            return (
              <a
                key={l.href}
                href={l.href}
                target="_blank"
                rel="noopener noreferrer"
                className="group rounded-xl border border-border bg-bg-elevated p-4 transition-colors hover:bg-bg-hover"
              >
                <div className="flex items-start justify-between">
                  <span className={"inline-flex h-9 w-9 items-center justify-center rounded-lg " + groupAccent[l.group]}>
                    <Icon className="h-4 w-4" />
                  </span>
                  <ExternalLink className="h-3.5 w-3.5 text-text-muted opacity-0 transition-opacity group-hover:opacity-100" />
                </div>
                <p className="mt-3 text-sm font-semibold text-foreground group-hover:text-accent">{l.title}</p>
                <p className="mt-0.5 text-xs text-text-muted">{l.description}</p>
                <p className="mt-2 truncate text-[11px] font-mono text-text-muted">
                  {l.href.replace(/^https?:\/\//, "")}
                </p>
              </a>
            );
          })}
        </div>
      </div>
    </section>
  );
}
`
}

func webProviders() string {
	return `"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            retry: 1,
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}
`
}

// viteAPIClient is the Vite/TanStack-flavoured API client. It uses
// import.meta.env.VITE_API_URL (the Vite convention) instead of
// process.env.NEXT_PUBLIC_API_URL (which is a Next.js-only prefix and
// undefined in a Vite project). Default is ” so the embedded single-binary
// deploy serves SPA + API from the same origin without configuration.
func viteAPIClient() string {
	return `import axios from "axios";

const API_URL = (import.meta.env.VITE_API_URL as string | undefined) ?? "";

export const api = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Auto-attach Idempotency-Key on unsafe methods so any mutation gets
// safe-retry semantics for free.
api.interceptors.request.use((config) => {
  const method = (config.method || "get").toUpperCase();
  const unsafe = method === "POST" || method === "PUT" || method === "PATCH" || method === "DELETE";
  if (unsafe && config.headers && !config.headers["Idempotency-Key"]) {
    config.headers["Idempotency-Key"] = crypto.randomUUID();
  }
  return config;
});
`
}

func webAPIClient() string {
	return `import axios from "axios";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const api = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
  // The browser attaches the HttpOnly grit_access / grit_refresh cookies
  // set by /api/auth/login automatically. Without this, axios skips them
  // on cross-origin requests in dev (api on :8080, web on :3000) and the
  // server treats every request as anonymous.
  withCredentials: true,
});

// Echo the grit_csrf cookie into X-CSRF-Token on every state-changing
// request. The cookie is intentionally not HttpOnly — it's the
// double-submit token, paired with the cookie the AutoCSRF middleware
// enforces. Safe-method requests don't need it; the middleware skips
// them and issues / refreshes the cookie as a side effect.
api.interceptors.request.use((config) => {
  if (typeof document !== "undefined") {
    const m = document.cookie.match(/(?:^|; )grit_csrf=([^;]+)/);
    if (m && config.headers) {
      config.headers["X-CSRF-Token"] = decodeURIComponent(m[1]);
    }
  }

  // Auto-attach Idempotency-Key on unsafe methods so any mutation gets
  // safe-retry semantics for free.
  const method = (config.method || "get").toUpperCase();
  const unsafe = method === "POST" || method === "PUT" || method === "PATCH" || method === "DELETE";
  if (unsafe && config.headers && !config.headers["Idempotency-Key"]) {
    config.headers["Idempotency-Key"] = crypto.randomUUID();
  }
  return config;
});

// v3.31.21: alias kept so generated React Query hooks that import
// { apiClient } from "@/lib/api" resolve symmetrically with apps/admin
// (which exports the same name from its own api-client.ts).
export const apiClient = api;
`
}

func webUseBlogsHook() string {
	return `"use client";

import { useQuery } from "@tanstack/react-query";
import type { Blog, PaginatedResponse } from "@repo/shared/types";
import { api } from "@/lib/api";

// Types live in packages/shared — never inline them in a hook. Reasons:
//   - the same Blog shape is consumed by web, admin, and the Go API (via
//     grit sync's generated counterpart). One source of truth.
//   - editing the model in Go and running 'grit sync' updates ALL
//     consumers in one shot; inline duplicates silently drift.
type BlogMeta = PaginatedResponse<Blog>["meta"];

export function usePublicBlogs(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ["public-blogs", page, pageSize],
    queryFn: async (): Promise<{ blogs: Blog[]; meta: BlogMeta | undefined }> => {
      const { data } = await api.get<PaginatedResponse<Blog>>(
        ` + "`" + `/api/blogs?page=${page}&page_size=${pageSize}` + "`" + `
      );
      return {
        blogs: data.data ?? [],
        meta: data.meta,
      };
    },
  });
}

export function usePublicBlog(slug: string) {
  return useQuery({
    queryKey: ["public-blog", slug],
    queryFn: async () => {
      const { data } = await api.get(` + "`" + `/api/blogs/${slug}` + "`" + `);
      return data.data as Blog;
    },
    enabled: !!slug,
  });
}

// Vite/TanStack blog routes import { useBlogs, useBlog } from this hook.
// Aliases so a single file works for both Next.js and Vite scaffolds.
export function useBlogs(page = 1, pageSize = 20) {
  const result = usePublicBlogs(page, pageSize);
  return { ...result, data: result.data?.blogs };
}

export const useBlog = usePublicBlog;
`
}

func webBlogListPage() string {
	return `"use client";

import { useState } from "react";
import Link from "next/link";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { usePublicBlogs } from "@/hooks/use-blogs";

export default function BlogListPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = usePublicBlogs(page, 9);
  const blogs = data?.blogs || [];
  const meta = data?.meta;

  return (
    <div className="mx-auto max-w-5xl px-6 py-16">
      {/* Header */}
      <div className="mb-12">
        <h1 className="text-4xl font-bold tracking-tight">Blog</h1>
        <p className="mt-2 text-text-secondary">
          Insights, tutorials, and updates from the team.
        </p>
      </div>

      {/* Blog grid */}
      {isLoading ? (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="rounded-xl border border-border bg-bg-elevated overflow-hidden animate-pulse"
            >
              <div className="h-52 bg-bg-hover" />
              <div className="p-5 space-y-3">
                <div className="h-3 bg-bg-hover rounded w-1/3" />
                <div className="h-5 bg-bg-hover rounded w-3/4" />
                <div className="h-3 bg-bg-hover rounded w-full" />
                <div className="h-3 bg-bg-hover rounded w-2/3" />
              </div>
            </div>
          ))}
        </div>
      ) : blogs.length > 0 ? (
        <>
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {blogs.map((blog) => (
              <Link
                key={blog.id}
                href={` + "`" + `/blog/${blog.slug}` + "`" + `}
                className="group rounded-xl border border-border bg-bg-elevated overflow-hidden hover:border-accent/40 hover:shadow-lg hover:shadow-accent/5 transition-all duration-300"
              >
                <div className="h-52 bg-bg-hover overflow-hidden">
                  {blog.image ? (
                    <img
                      src={blog.image}
                      alt={blog.title}
                      className="h-full w-full object-cover group-hover:scale-105 transition-transform duration-500"
                    />
                  ) : (
                    <div className="h-full w-full flex items-center justify-center bg-gradient-to-br from-accent/10 to-accent/5">
                      <span className="text-5xl font-bold text-accent/20">
                        {blog.title.charAt(0)}
                      </span>
                    </div>
                  )}
                </div>
                <div className="p-5">
                  <p className="text-xs text-text-muted mb-2.5">
                    {new Date(
                      blog.published_at || blog.created_at
                    ).toLocaleDateString("en-US", {
                      month: "short",
                      day: "numeric",
                      year: "numeric",
                    })}
                  </p>
                  <h2 className="font-semibold text-foreground group-hover:text-accent transition-colors line-clamp-2 text-lg leading-snug">
                    {blog.title}
                  </h2>
                  {blog.excerpt && (
                    <p className="mt-2.5 text-sm text-text-secondary line-clamp-3 leading-relaxed">
                      {blog.excerpt}
                    </p>
                  )}
                  <span className="mt-4 inline-block text-xs font-medium text-accent group-hover:text-accent-hover transition-colors">
                    Read more &rarr;
                  </span>
                </div>
              </Link>
            ))}
          </div>

          {/* Pagination */}
          {meta && meta.pages > 1 && (
            <div className="mt-12 flex items-center justify-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="flex items-center gap-1 rounded-lg border border-border bg-bg-elevated px-3 py-2 text-sm text-text-secondary hover:bg-bg-hover hover:text-foreground transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              >
                <ChevronLeft className="h-4 w-4" />
                Previous
              </button>
              <div className="flex items-center gap-1 px-3">
                {Array.from({ length: meta.pages }).map((_, i) => (
                  <button
                    key={i + 1}
                    onClick={() => setPage(i + 1)}
                    className={` + "`" + `h-8 w-8 rounded-lg text-sm font-medium transition-colors ${
                      page === i + 1
                        ? "bg-accent text-white"
                        : "text-text-secondary hover:bg-bg-hover hover:text-foreground"
                    }` + "`" + `}
                  >
                    {i + 1}
                  </button>
                ))}
              </div>
              <button
                onClick={() => setPage((p) => Math.min(meta.pages, p + 1))}
                disabled={page >= meta.pages}
                className="flex items-center gap-1 rounded-lg border border-border bg-bg-elevated px-3 py-2 text-sm text-text-secondary hover:bg-bg-hover hover:text-foreground transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Next
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-20">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-bg-elevated border border-border">
            <span className="text-2xl text-text-muted">&#9998;</span>
          </div>
          <h3 className="text-lg font-semibold text-foreground">No posts yet</h3>
          <p className="mt-1 text-sm text-text-muted">
            Blog posts will appear here once published from the admin panel.
          </p>
        </div>
      )}
    </div>
  );
}
`
}

func webBlogDetailPage() string {
	return `"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Calendar } from "lucide-react";
import { usePublicBlog } from "@/hooks/use-blogs";

export default function BlogDetailPage() {
  const params = useParams();
  const slug = typeof params.slug === "string" ? params.slug : "";
  const { data: blog, isLoading, error } = usePublicBlog(slug);

  if (isLoading) {
    return (
      <div className="mx-auto max-w-3xl px-6 py-16 animate-pulse">
        <div className="h-4 bg-bg-hover rounded w-24 mb-8" />
        <div className="h-8 bg-bg-hover rounded w-3/4 mb-4" />
        <div className="h-4 bg-bg-hover rounded w-1/3 mb-12" />
        <div className="aspect-[2/1] bg-bg-hover rounded-xl mb-12" />
        <div className="space-y-4">
          <div className="h-4 bg-bg-hover rounded w-full" />
          <div className="h-4 bg-bg-hover rounded w-full" />
          <div className="h-4 bg-bg-hover rounded w-5/6" />
          <div className="h-4 bg-bg-hover rounded w-full" />
          <div className="h-4 bg-bg-hover rounded w-4/6" />
        </div>
      </div>
    );
  }

  if (error || !blog) {
    return (
      <div className="mx-auto max-w-3xl px-6 py-16 text-center">
        <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-bg-elevated border border-border">
          <span className="text-2xl text-text-muted">404</span>
        </div>
        <h1 className="text-xl font-semibold text-foreground">Post not found</h1>
        <p className="mt-2 text-sm text-text-muted">
          The blog post you're looking for doesn't exist or has been removed.
        </p>
        <Link
          href="/blog"
          className="mt-6 inline-flex items-center gap-2 rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent-hover transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Blog
        </Link>
      </div>
    );
  }

  return (
    <article className="mx-auto max-w-3xl px-6 py-16">
      {/* Back link */}
      <Link
        href="/blog"
        className="inline-flex items-center gap-1.5 text-sm text-text-secondary hover:text-foreground transition-colors mb-8"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Blog
      </Link>

      {/* Title and meta */}
      <header className="mb-10">
        <h1 className="text-3xl sm:text-4xl font-bold tracking-tight leading-tight">
          {blog.title}
        </h1>
        <div className="mt-4 flex items-center gap-2 text-sm text-text-muted">
          <Calendar className="h-4 w-4" />
          <time dateTime={blog.published_at || blog.created_at}>
            {new Date(blog.published_at || blog.created_at).toLocaleDateString(
              "en-US",
              {
                month: "long",
                day: "numeric",
                year: "numeric",
              }
            )}
          </time>
        </div>
      </header>

      {/* Cover image */}
      {blog.image && (
        <div className="mb-12 rounded-xl overflow-hidden border border-border">
          <img
            src={blog.image}
            alt={blog.title}
            className="w-full h-auto object-cover"
          />
        </div>
      )}

      {/* Content */}
      <div
        className="prose-blog"
        dangerouslySetInnerHTML={{ __html: blog.content }}
      />

      {/* Bottom nav */}
      <div className="mt-16 pt-8 border-t border-border/50">
        <Link
          href="/blog"
          className="inline-flex items-center gap-1.5 text-sm text-accent hover:text-accent-hover transition-colors font-medium"
        >
          <ArrowLeft className="h-4 w-4" />
          All posts
        </Link>
      </div>
    </article>
  );
}
`
}

func webErrorPage() string {
	return `"use client";

import { useEffect } from "react";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="flex min-h-[60vh] items-center justify-center px-4">
      <div className="w-full max-w-md text-center">
        <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-full bg-red-500/10 border border-red-500/20">
          <svg className="h-8 w-8 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>
        <h2 className="mb-2 text-2xl font-bold text-foreground">Something went wrong</h2>
        <p className="mb-6 text-muted-foreground">
          An unexpected error occurred. You can try again or go back.
        </p>
        {error.digest && (
          <p className="mb-4 text-xs text-muted-foreground/60 font-mono">Error ID: {error.digest}</p>
        )}
        <div className="flex gap-3 justify-center">
          <button
            onClick={() => window.history.back()}
            className="inline-flex items-center gap-2 rounded-lg border border-border px-4 py-2.5 text-sm font-medium text-foreground hover:bg-accent/50 transition-colors"
          >
            Go Back
          </button>
          <button
            onClick={reset}
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            Try Again
          </button>
        </div>
      </div>
    </div>
  );
}
`
}

func webNotFoundPage() string {
	return `import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex min-h-[60vh] items-center justify-center px-4">
      <div className="w-full max-w-md text-center">
        <p className="mb-4 text-7xl font-bold text-primary">404</p>
        <h2 className="mb-2 text-2xl font-bold text-foreground">Page not found</h2>
        <p className="mb-8 text-muted-foreground">
          The page you&apos;re looking for doesn&apos;t exist or has been moved.
        </p>
        <div className="flex gap-3 justify-center">
          <Link
            href="/"
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            Home
          </Link>
        </div>
      </div>
    </div>
  );
}
`
}

func webGlobalErrorPage() string {
	return `"use client";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="en" className="dark">
      <body style={{ minHeight: "100vh", backgroundColor: "#0a0a0f", fontFamily: "system-ui, sans-serif", margin: 0 }}>
        <div style={{ display: "flex", minHeight: "100vh", alignItems: "center", justifyContent: "center", padding: "1rem" }}>
          <div style={{ maxWidth: "28rem", textAlign: "center" }}>
            <div style={{ margin: "0 auto 1.5rem", display: "flex", height: "4rem", width: "4rem", alignItems: "center", justifyContent: "center", borderRadius: "9999px", backgroundColor: "rgba(239,68,68,0.1)", border: "1px solid rgba(239,68,68,0.2)" }}>
              <svg style={{ height: "2rem", width: "2rem", color: "#f87171" }} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h2 style={{ marginBottom: "0.5rem", fontSize: "1.5rem", fontWeight: 700, color: "#e8e8f0" }}>Application Error</h2>
            <p style={{ marginBottom: "1.5rem", color: "#9090a8" }}>
              A critical error occurred. Please try refreshing the page.
            </p>
            {error.digest && (
              <p style={{ marginBottom: "1rem", fontSize: "0.75rem", color: "#606078", fontFamily: "monospace" }}>Error ID: {error.digest}</p>
            )}
            <button
              onClick={reset}
              style={{ display: "inline-flex", alignItems: "center", gap: "0.5rem", borderRadius: "0.5rem", backgroundColor: "#6c5ce7", padding: "0.625rem 1.25rem", fontSize: "0.875rem", fontWeight: 500, color: "white", border: "none", cursor: "pointer" }}
            >
              Try Again
            </button>
          </div>
        </div>
      </body>
    </html>
  );
}
`
}

// ---------------------------------------------------------------------------
// Web Auth Pages
// ---------------------------------------------------------------------------




func webAuthCallback() string {
	return `"use client";

import { Suspense, useEffect, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Cookies from "js-cookie";
import { api } from "@/lib/api";

// useSearchParams forces a client bailout during prerender unless it's
// wrapped in a Suspense boundary. The inner component does the work;
// the page export just provides the boundary.
export default function AuthCallbackPage() {
  return (
    <Suspense fallback={<CallbackSpinner />}>
      <CallbackInner />
    </Suspense>
  );
}

function CallbackInner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const processed = useRef(false);

  useEffect(() => {
    if (processed.current) return;
    processed.current = true;

    const accessToken = searchParams.get("access_token");
    const refreshToken = searchParams.get("refresh_token");
    const error = searchParams.get("error");

    if (error) {
      router.push("/login?error=" + encodeURIComponent(error));
      return;
    }

    if (accessToken && refreshToken) {
      Cookies.set("access_token", accessToken, { expires: 1 });
      Cookies.set("refresh_token", refreshToken, { expires: 7 });

      api
        .get("/api/auth/me", {
          headers: { Authorization: "Bearer " + accessToken },
        })
        .then(() => {
          router.push("/");
        })
        .catch(() => {
          router.push("/");
        });
    } else {
      router.push("/login?error=Authentication+failed");
    }
  }, [searchParams, router]);

  return <CallbackSpinner />;
}

function CallbackSpinner() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="text-center">
        <div className="inline-flex h-10 w-10 animate-spin items-center justify-center rounded-full border-2 border-accent border-t-transparent" />
        <p className="mt-4 text-sm text-text-secondary">Signing you in...</p>
      </div>
    </div>
  );
}
`
}

func webUseAuth() string {
	return `import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import type {
  User,
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  ApiResponse,
} from "@repo/shared/types";
import { api } from "@/lib/api";
import { setWebSessionMarker, clearWebSessionMarker } from "@/lib/web-session";

// Token storage policy (Grit 3.26+):
//   - The API issues HttpOnly cookies (grit_access + grit_refresh) on
//     login/register/refresh and clears them on logout. The browser
//     handles them automatically — JS never reads or writes the access
//     token, so XSS can't exfiltrate it.
//   - The axios client uses withCredentials: true so the browser
//     attaches the cookies on every request.
//   - There is no Bearer-header path here. Mobile/desktop bearer
//     clients live in their own apps; the web app is cookie-only.
//
// Types live in packages/shared/types/user.ts — they're consumed by
// web, admin, and the Go API (via grit sync). Never inline them.

export function useMe() {
  return useQuery<User | null>({
    queryKey: ["me"],
    queryFn: async () => {
      try {
        const { data } = await api.get<ApiResponse<User>>("/api/auth/me");
        return data.data;
      } catch (err: unknown) {
        // 401 is the canonical "not logged in" — return null instead
        // of throwing so guarded pages can read user === null cleanly.
        const e = err as { response?: { status?: number } };
        if (e.response?.status === 401) return null;
        throw err;
      }
    },
    retry: false,
    staleTime: 10 * 60 * 1000,
  });
}

export function useAuth() {
  const { data: user, isLoading, isError } = useMe();
  return {
    user: user ?? null,
    isAuthenticated: !!user,
    isLoading,
    isError,
  };
}

export function useLogin() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (credentials: LoginRequest) => {
      // POST is enough — the API sets HttpOnly cookies on the response.
      // The 'tokens' field is still in the JSON body so native bearer
      // clients work too, but the browser ignores it.
      const { data } = await api.post<ApiResponse<AuthResponse>>(
        "/api/auth/login",
        credentials
      );
      return data;
    },
    onSuccess: (data) => {
      queryClient.setQueryData(["me"], data.data.user);
      // v3.31.42: stamp the web-origin marker so the middleware
      // doesn't bounce signed-in web users to /login on the next
      // navigation. See lib/web-session.ts for the rationale.
      setWebSessionMarker();
      router.push("/");
    },
  });
}

export function useRegister() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: RegisterRequest) => {
      const { data } = await api.post<ApiResponse<AuthResponse>>(
        "/api/auth/register",
        payload
      );
      return data;
    },
    onSuccess: (data) => {
      queryClient.setQueryData(["me"], data.data.user);
      // v3.31.42: same web-origin marker as useLogin.
      setWebSessionMarker();
      router.push("/");
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      try {
        await api.post("/api/auth/logout");
      } catch {
        // The cookies are already cleared by the API's Set-Cookie even
        // on a 4xx; just make sure local state is wiped either way.
      }
    },
    onSettled: () => {
      queryClient.clear();
      // v3.31.42: clear the marker so middleware no longer sees
      // a session on the next request.
      clearWebSessionMarker();
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
    },
  });
}
`
}

func webAuthProvider() string {
	return `"use client";

import { createContext, useContext } from "react";
import type { User } from "@repo/shared/types";
import { useMe } from "@/hooks/use-auth";

// User shape comes from packages/shared/types/user.ts — same shape the
// Go API serialises, same shape the admin imports. Never inline it.

interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextValue>({
  user: null,
  isAuthenticated: false,
  isLoading: true,
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data: user, isLoading } = useMe();

  return (
    <AuthContext.Provider
      value={{
        user: user ?? null,
        isAuthenticated: !!user,
        isLoading,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuthContext() {
  return useContext(AuthContext);
}
`
}

// webPublicFormPage — public-facing page at /forms/[token] that
// renders a generic submission form for any FormShare-exposed resource.
//
// Designed to be minimal but functional:
//   - Fetches /api/public/forms/<token> to confirm the link works +
//     learn the resource_name and whether a password is required.
//   - If password required, shows a gate input. Submitting hits the
//     submit endpoint with the password+empty fields to probe; on a
//     successful probe we move to the form. (Phase 2 hardens this by
//     adding a dedicated /check-password endpoint that issues a
//     short-lived cookie — for v1 the password is sent with each
//     submit so the page itself stays stateless.)
//   - The form itself is a one-size-fits-all key/value editor where
//     visitors fill labelled inputs. The shape is hard-coded to a
//     name+email+message pattern that covers the dominant case (lead
//     forms, contact forms, applications). Resources with different
//     shapes are best exposed via `grit expose form` once Phase 3
//     lands — that command generates a page tailored to the resource's
//     actual fields.
func webPublicFormPage() string {
	return `"use client";

// v3.31.43+: ShareInfo carries the resource's actual field shape
// (built server-side by services.PublicFields). v3.31.50 adds
// operator-customised title + description and respects the
// hidden_fields list -- so this page no longer renders a hardcoded
// name/email/phone/message contact form. Whatever fields the
// resource's struct tags expose are what the visitor sees.

import { useEffect, useState, use } from "react";
import axios from "axios";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface PageProps {
  params: Promise<{ token: string }>;
}

type PublicFieldType =
  | "text"
  | "email"
  | "tel"
  | "textarea"
  | "number"
  | "checkbox"
  | "date"
  | "datetime"
  | "file";

interface PublicField {
  key: string;
  label: string;
  type: PublicFieldType;
  required: boolean;
}

interface ShareInfo {
  resource_name: string;
  has_password: boolean;
  label: string;
  custom_title: string;
  custom_description: string;
  fields: PublicField[];
}

export default function PublicFormPage({ params }: PageProps) {
  const { token } = use(params);
  const [info, setInfo] = useState<ShareInfo | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    axios.get(API_URL + "/api/public/forms/" + token)
      .then((res) => setInfo(res.data.data))
      .catch((err) => {
        setError(err?.response?.data?.error?.message || "Link not found or disabled");
      })
      .finally(() => setLoading(false));
  }, [token]);

  if (loading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-50 p-4">
        <div className="text-sm text-slate-500">Loading…</div>
      </main>
    );
  }

  if (error || !info) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-50 p-4">
        <div className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm">
          <h1 className="text-xl font-semibold text-slate-900">Link unavailable</h1>
          <p className="mt-2 text-sm text-slate-500">{error ?? "Unknown error"}</p>
        </div>
      </main>
    );
  }

  // v3.31.50 -- title falls back through three sources:
  //   1. operator-set custom_title (best)
  //   2. operator-set label (legacy -- pre-v3.31.50 shares)
  //   3. resource name (worst -- bare default)
  const title =
    info.custom_title?.trim() ||
    info.label?.trim() ||
    info.resource_name + " submission";
  const description =
    info.custom_description?.trim() ||
    "Fill out the form below to submit a new " + info.resource_name + ".";

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-50 p-4">
      <div className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
        <h1 className="text-2xl font-semibold text-slate-900">{title}</h1>
        <p className="mt-1 text-sm text-slate-500">{description}</p>

        <PublicForm token={token} info={info} />
      </div>
    </main>
  );
}

interface PublicFormProps {
  token: string;
  info: ShareInfo;
}

function PublicForm({ token, info }: PublicFormProps) {
  const [password, setPassword] = useState("");
  // Mixed value types so checkbox + number fields can survive the
  // round-trip without coercion ceremony at submit time.
  const [fields, setFields] = useState<Record<string, string | number | boolean>>(() => {
    const initial: Record<string, string | number | boolean> = {};
    for (const f of info.fields) {
      initial[f.key] = f.type === "checkbox" ? false : "";
    }
    return initial;
  });
  const [submitting, setSubmitting] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const update = (key: string, value: string | number | boolean) => {
    setFields((prev) => ({ ...prev, [key]: value }));
  };

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      // Strip file fields -- not supported on public shares yet
      // (auth-gated /api/uploads endpoint). Sending them would
      // confuse the dispatcher's typed unmarshal.
      const payload: Record<string, string | number | boolean> = {};
      for (const f of info.fields) {
        if (f.type === "file") continue;
        payload[f.key] = fields[f.key];
      }
      await axios.post(API_URL + "/api/public/forms/" + token + "/submit", {
        _password: password,
        fields: payload,
      });
      setDone(true);
    } catch (err) {
      const e = err as { response?: { data?: { error?: { message?: string } } } };
      setError(e?.response?.data?.error?.message || "Submission failed");
    } finally {
      setSubmitting(false);
    }
  };

  if (done) {
    return (
      <div className="mt-6 rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-700">
        <p className="font-medium">Thank you</p>
        <p className="mt-1">Your submission was received.</p>
      </div>
    );
  }

  return (
    <form onSubmit={onSubmit} className="mt-6 space-y-4">
      {info.has_password && (
        <Field
          field={{ key: "_password", label: "Password", type: "text", required: true }}
          inputType="password"
          value={password}
          onChange={(v) => setPassword(String(v))}
          hint="This form is password-protected — ask whoever shared the link."
        />
      )}

      {info.fields.length === 0 && (
        <div className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-700">
          This resource has no public-form fields registered on the API.
          Ask the operator to re-generate the resource so the form-share
          dispatcher picks up the field schema.
        </div>
      )}

      {info.fields.map((f) => (
        <Field
          key={f.key}
          field={f}
          value={fields[f.key] ?? (f.type === "checkbox" ? false : "")}
          onChange={(v) => update(f.key, v)}
        />
      ))}

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </div>
      )}

      <button
        type="submit"
        disabled={submitting}
        className="w-full rounded-lg bg-slate-900 px-4 py-2.5 text-sm font-semibold text-white hover:bg-slate-800 disabled:opacity-50"
      >
        {submitting ? "Sending…" : "Submit"}
      </button>
    </form>
  );
}

interface FieldProps {
  field: PublicField;
  value: string | number | boolean;
  onChange: (v: string | number | boolean) => void;
  inputType?: string;
  hint?: string;
}

function Field({ field, value, onChange, inputType, hint }: FieldProps) {
  const labelClass = "block text-sm font-medium text-slate-700";
  const inputClass =
    "block w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-900/10";

  if (field.type === "file") {
    return (
      <div className="space-y-1.5">
        <label className={labelClass}>{field.label}</label>
        <div className="rounded-lg border border-dashed border-slate-300 bg-slate-50 px-3 py-3 text-xs text-slate-500">
          File uploads aren&apos;t supported on public-share forms.
          {field.required
            ? " The operator must collect this file through a different channel."
            : " You can leave this blank."}
        </div>
      </div>
    );
  }

  if (field.type === "textarea") {
    return (
      <div className="space-y-1.5">
        <label className={labelClass}>
          {field.label}
          {field.required && <span className="ml-1 text-red-500">*</span>}
        </label>
        <textarea
          value={String(value)}
          onChange={(e) => onChange(e.target.value)}
          required={field.required}
          rows={4}
          className={inputClass}
        />
        {hint && <p className="text-xs text-slate-500">{hint}</p>}
      </div>
    );
  }

  if (field.type === "checkbox") {
    return (
      <div className="space-y-1.5">
        <label className="flex items-center gap-2 text-sm font-medium text-slate-700">
          <input
            type="checkbox"
            checked={Boolean(value)}
            onChange={(e) => onChange(e.target.checked)}
            className="h-4 w-4 rounded border-slate-300"
          />
          {field.label}
          {field.required && <span className="ml-1 text-red-500">*</span>}
        </label>
        {hint && <p className="text-xs text-slate-500">{hint}</p>}
      </div>
    );
  }

  if (field.type === "number") {
    return (
      <div className="space-y-1.5">
        <label className={labelClass}>
          {field.label}
          {field.required && <span className="ml-1 text-red-500">*</span>}
        </label>
        <input
          type="number"
          value={value === "" ? "" : String(value)}
          onChange={(e) => onChange(e.target.value === "" ? "" : Number(e.target.value))}
          required={field.required}
          className={inputClass}
        />
        {hint && <p className="text-xs text-slate-500">{hint}</p>}
      </div>
    );
  }

  const htmlType =
    inputType ??
    (field.type === "email"
      ? "email"
      : field.type === "tel"
        ? "tel"
        : field.type === "date"
          ? "date"
          : field.type === "datetime"
            ? "datetime-local"
            : "text");

  return (
    <div className="space-y-1.5">
      <label className={labelClass}>
        {field.label}
        {field.required && <span className="ml-1 text-red-500">*</span>}
      </label>
      <input
        type={htmlType}
        value={String(value)}
        onChange={(e) => onChange(e.target.value)}
        required={field.required}
        className={inputClass}
      />
      {hint && <p className="text-xs text-slate-500">{hint}</p>}
    </div>
  );
}
`
}
