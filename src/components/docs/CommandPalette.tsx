import { useEffect, useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import {
  BookOpen,
  Flag,
  FileText,
  Search,
} from "lucide-react";
import { commands } from "@/data/commands";

interface FlagEntry {
  flag: string;
  description: string;
  command: string;
}

const pages = [
  { title: "Home", url: "/" },
  { title: "Commands", url: "/commands" },
  { title: "Getting Started", url: "/getting-started" },
  { title: "Configuration", url: "/config" },
  { title: "Architecture", url: "/architecture" },
  { title: "Watch", url: "/watch" },
  { title: "Release", url: "/release" },
  { title: "GoMod", url: "/gomod" },
  { title: "Projects", url: "/projects" },
  { title: "Makefile", url: "/makefile" },
  { title: "History", url: "/history" },
  { title: "Stats", url: "/stats" },
  { title: "Detection", url: "/project-detection" },
  { title: "Generic CLI", url: "/generic-cli" },
  { title: "Changelog", url: "/changelog" },
  { title: "Flag Reference", url: "/flags" },
  { title: "Interactive Examples", url: "/examples" },
  { title: "Interactive TUI", url: "/interactive" },
  { title: "Batch Actions", url: "/batch-actions" },
  { title: "Clear Release JSON", url: "/clear-release-json" },
  { title: "Bookmarks", url: "/bookmarks" },
  { title: "Export", url: "/export" },
  { title: "Import", url: "/import" },
  { title: "Profile", url: "/profile" },
  { title: "Diff Profiles", url: "/diff-profiles" },
  { title: "Spec Index", url: "/spec" },
  { title: "Spec: scan all", url: "/scan-all" },
  { title: "Spec: desktop-sync (ds = gd)", url: "/desktop-sync" },
  { title: "Spec: github-desktop (gd)", url: "/github-desktop" },
  { title: "Spec: scan gd (bulk)", url: "/scan-gd" },
  { title: "Spec: clone multi-URL", url: "/clone-multi" },
];

const CommandPalette = () => {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((o) => !o);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  const allFlags = useMemo<FlagEntry[]>(() => {
    const rows: FlagEntry[] = [];
    for (const cmd of commands) {
      if (!cmd.flags) continue;
      for (const f of cmd.flags) {
        rows.push({ flag: f.flag, description: f.description, command: cmd.name });
      }
    }
    return rows;
  }, []);

  const go = (url: string) => {
    setOpen(false);
    navigate(url);
  };

  return (
    <>
      <Tooltip>
        <TooltipTrigger asChild>
          <button
            onClick={() => setOpen(true)}
            aria-label="Open command palette (search commands, flags, pages)"
            className="flex items-center gap-2 px-3 py-1.5 rounded-lg border border-border bg-card text-muted-foreground text-xs font-sans hover:bg-muted/50 hover:text-foreground transition-colors"
          >
            <Search className="h-3 w-3" />
            <span className="hidden sm:inline">Search...</span>
            <kbd className="hidden sm:inline-flex items-center gap-0.5 rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] font-mono text-muted-foreground">
              ⌘K
            </kbd>
          </button>
        </TooltipTrigger>
        <TooltipContent side="bottom">Search commands, flags & pages (⌘K)</TooltipContent>
      </Tooltip>

      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Search commands, flags, pages..." />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>

          <CommandGroup heading="Pages">
            {pages.map((page) => (
              <CommandItem key={page.url} onSelect={() => go(page.url)}>
                <FileText className="mr-2 h-4 w-4 text-muted-foreground" />
                <span>{page.title}</span>
              </CommandItem>
            ))}
          </CommandGroup>

          <CommandSeparator />

          <CommandGroup heading="Commands">
            {commands.map((cmd) => (
              <CommandItem key={cmd.name} onSelect={() => go("/commands")} keywords={[cmd.name, cmd.alias ?? "", cmd.description]}>
                <BookOpen className="mr-2 h-4 w-4 text-muted-foreground" />
                <span className="font-sans font-semibold">{cmd.name}</span>
                {cmd.alias && (
                  <span className="ml-1 text-xs font-sans font-medium text-foreground bg-primary/10 border border-primary/20 px-1.5 py-0.5 rounded dark:bg-primary/15 dark:text-primary">{cmd.alias}</span>
                )}
                <span className="text-muted-foreground ml-2 text-xs truncate">{cmd.description}</span>
              </CommandItem>
            ))}
          </CommandGroup>

          <CommandSeparator />

          <CommandGroup heading="Flags">
            {allFlags.map((f, i) => (
              <CommandItem key={`${f.command}-${f.flag}-${i}`} onSelect={() => go("/flags")} keywords={[f.flag, f.description, f.command]}>
                <Flag className="mr-2 h-4 w-4 text-muted-foreground" />
                <span className="font-mono text-primary">{f.flag}</span>
                <span className="text-muted-foreground ml-2 text-xs font-sans">{f.command}</span>
                <span className="text-muted-foreground ml-1 text-xs truncate hidden sm:inline font-sans">— {f.description}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </>
  );
};

export default CommandPalette;
