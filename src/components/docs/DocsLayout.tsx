import { useEffect, useState } from "react";
import { Sun, Moon } from "lucide-react";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { DocsSidebar } from "@/components/docs/DocsSidebar";
import CommandPalette from "@/components/docs/CommandPalette";
import { VERSION } from "@/constants/index";
import { getCurrentTheme, setTheme } from "@/lib/theme";

interface DocsLayoutProps {
  children: React.ReactNode;
}

const DocsLayout = ({ children }: DocsLayoutProps) => {
  const [dark, setDark] = useState(() => getCurrentTheme() === "dark");

  useEffect(() => {
    setTheme(dark ? "dark" : "light");
  }, [dark]);

  return (
    <SidebarProvider>
      <div className="min-h-screen flex w-full bg-background text-foreground">
        <DocsSidebar />
        <div className="flex-1 flex flex-col min-w-0">
          <header className="sticky top-0 z-10 flex h-12 items-center border-b border-sidebar-border bg-sidebar/95 backdrop-blur-sm">
            <SidebarTrigger className="ml-3 rounded-sm border border-sidebar-border bg-sidebar-accent/60 text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground" />
            <span className="ml-3 text-[11px] font-mono uppercase tracking-[0.16em] text-muted-foreground">Explorer</span>
            <span className="ml-3 text-sm font-mono text-foreground">gitmap docs</span>
            <span className="ml-2 rounded-sm border border-border bg-card px-2 py-0.5 text-[11px] font-mono text-muted-foreground shadow-sm">
              {VERSION}
            </span>
            <div
              role="radiogroup"
              aria-label="Color theme"
              className="ml-2 inline-flex items-center rounded-sm border border-border bg-card p-0.5 shadow-sm"
            >
              <button
                type="button"
                role="radio"
                aria-checked={dark}
                onClick={() => setDark(true)}
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
                onClick={() => setDark(false)}
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
            <div className="ml-auto mr-3">
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
