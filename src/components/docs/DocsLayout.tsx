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
            <button
              type="button"
              onClick={() => setDark((d) => !d)}
              aria-label={dark ? "Switch to light mode" : "Switch to dark mode"}
              title={dark ? "Switch to light mode" : "Switch to dark mode"}
              className="ml-2 inline-flex h-6 w-6 items-center justify-center rounded-sm border border-border bg-card text-muted-foreground transition-colors duration-300 hover:bg-secondary hover:text-foreground"
            >
              {dark ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
            </button>
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
