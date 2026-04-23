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
      <div className="min-h-screen flex w-full">
        <DocsSidebar />
        <div className="flex-1 flex flex-col min-w-0">
          <header className="h-12 flex items-center border-b border-border sticky top-0 bg-background/80 backdrop-blur-sm z-10">
            <SidebarTrigger className="ml-3" />
            <span className="ml-3 text-sm font-mono text-muted-foreground">gitmap documentation</span>
            <span className="ml-2 px-2 py-0.5 text-xs font-mono bg-primary/10 text-foreground border border-primary/20 rounded transition-colors duration-300 hover:border-primary/40 hover:shadow-sm hover:shadow-primary/10 dark:bg-primary/20 dark:text-primary dark:border-primary/40 dark:hover:border-primary/60">
              {VERSION}
            </span>
            <button
              type="button"
              onClick={() => setDark((d) => !d)}
              aria-label={dark ? "Switch to light mode" : "Switch to dark mode"}
              title={dark ? "Switch to light mode" : "Switch to dark mode"}
              className="ml-2 inline-flex items-center justify-center h-6 w-6 rounded border border-primary/20 bg-primary/10 text-foreground hover:bg-primary/20 hover:border-primary/40 hover:shadow-sm hover:shadow-primary/10 dark:bg-primary/20 dark:text-primary dark:border-primary/40 dark:hover:bg-primary/30 transition-colors duration-300"
            >
              {dark ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
            </button>
            <div className="ml-auto mr-3">
              <CommandPalette />
            </div>
          </header>
          <main className="flex-1 overflow-auto">
            <div className="max-w-4xl mx-auto px-6 py-8">
              {children}
            </div>
          </main>
        </div>
      </div>
    </SidebarProvider>
  );
};

export default DocsLayout;
