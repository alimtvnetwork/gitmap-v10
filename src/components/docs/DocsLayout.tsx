import { Sun, Moon } from "lucide-react";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { DocsSidebar } from "@/components/docs/DocsSidebar";
import CommandPalette from "@/components/docs/CommandPalette";
import CopyPaletteButton from "@/components/docs/CopyPaletteButton";
import { VERSION } from "@/constants/index";
import { useTheme } from "@/hooks/useTheme";

interface DocsLayoutProps {
  children: React.ReactNode;
}

const DocsLayout = ({ children }: DocsLayoutProps) => {
  const { isDark, isSystem, setTheme } = useTheme();
  const dark = isDark;

  return (
    <SidebarProvider>
      <div className="h-screen flex w-full overflow-hidden bg-background text-foreground">
        <DocsSidebar />
        <div className="flex-1 flex flex-col min-w-0 min-h-0">
          <header className="sticky top-0 z-10 flex shrink-0 flex-wrap items-center gap-2 border-b border-sidebar-border bg-sidebar/95 px-3 py-2 backdrop-blur-sm">
            <SidebarTrigger className="shrink-0 rounded-sm border border-sidebar-border bg-sidebar-accent/60 text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground" />
            <span className="shrink-0 text-[11px] font-mono uppercase tracking-[0.16em] text-muted-foreground">Explorer</span>
            <span className="shrink-0 text-sm font-mono text-foreground">gitmap docs</span>
            <span className="shrink-0 rounded-sm border border-border bg-card px-2 py-0.5 text-[11px] font-mono text-muted-foreground shadow-sm">
              {VERSION}
            </span>
            <div
              role="radiogroup"
              aria-label="Color theme"
              className="inline-flex shrink-0 items-center rounded-sm border border-border bg-card p-0.5 shadow-sm"
            >
              <button
                type="button"
                role="radio"
                aria-checked={dark}
                onClick={() => setTheme("dark")}
                className={[
                  "inline-flex items-center gap-1.5 rounded-[3px] px-2 py-0.5 text-[11px] font-sans font-medium transition-colors duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 focus-visible:ring-offset-card",
                  dark
                    ? "bg-secondary text-foreground"
                    : "text-muted-foreground hover:text-foreground",
                ].join(" ")}
              >
                <Moon className="h-3 w-3" aria-hidden="true" />
                Dark
              </button>
              <button
                type="button"
                role="radio"
                aria-checked={!dark}
                onClick={() => setTheme("light")}
                className={[
                  "inline-flex items-center gap-1.5 rounded-[3px] px-2 py-0.5 text-[11px] font-sans font-medium transition-colors duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 focus-visible:ring-offset-card",
                  !dark
                    ? "bg-secondary text-foreground"
                    : "text-muted-foreground hover:text-foreground",
                ].join(" ")}
              >
                <Sun className="h-3 w-3" aria-hidden="true" />
                Light
              </button>
            </div>
            <span
              className="hidden shrink-0 text-[11px] font-mono text-muted-foreground lg:inline"
              aria-live="polite"
            >
              {dark ? "VS Code Dark+" : "VS Code Light+"}
            </span>
            {isSystem && (
              <span
                className="hidden shrink-0 rounded-sm border border-border bg-card px-1.5 py-0.5 text-[10px] font-mono uppercase tracking-[0.12em] text-muted-foreground shadow-sm lg:inline"
                title="Following OS prefers-color-scheme — pick Dark or Light to override"
              >
                System
              </span>
            )}
            <div className="shrink-0">
              <CopyPaletteButton />
            </div>
            <div className="ml-auto shrink-0">
              <CommandPalette />
            </div>
          </header>
          <main className="flex-1 overflow-auto bg-background">
            <div className="mx-auto max-w-5xl px-6 py-8">
              {children}
            </div>
          </main>
        </div>
      </div>
    </SidebarProvider>
  );
};

export default DocsLayout;
