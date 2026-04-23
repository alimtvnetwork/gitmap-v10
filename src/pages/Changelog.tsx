import DocsLayout from "@/components/docs/DocsLayout";
import TerminalDemo from "@/components/docs/TerminalDemo";
import { changelog } from "@/data/changelog";
import { Tag, ChevronDown, ChevronRight } from "lucide-react";
import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

const terminalLines = [
  { text: "gitmap list-versions", type: "input" as const, delay: 800 },
  { text: "", type: "output" as const },
  { text: "  VERSION    CHANGES  HIGHLIGHTS", type: "header" as const },
  { text: "  ───────    ───────  ──────────", type: "output" as const },
  { text: "  v2.52.0       12   Lock detection, install system, uninstall", type: "accent" as const },
  { text: "  v2.51.0        8   Help reorganization, --compact flag", type: "output" as const },
  { text: "  v2.48.1        2   clone-next auto-cd into new directory", type: "output" as const },
  { text: "  v2.48.0        4   list-releases tag discovery, upsert", type: "output" as const },
  { text: "  v2.36.7        8   Integration tests, TUI temp releases", type: "output" as const },
  { text: "  v2.36.6       16   14 file splits, 200-line compliance", type: "output" as const },
  { text: "", type: "output" as const },
  { text: "  6 versions shown · 50 total changes", type: "accent" as const },
  { text: "", type: "output" as const },
  { text: "gitmap changelog v2.52.0", type: "input" as const, delay: 1000 },
  { text: "", type: "output" as const },
  { text: "  ## v2.52.0", type: "header" as const },
  { text: "  • Lock detection for clone-next via handle.exe/lsof", type: "output" as const },
  { text: "  • Install system with SQLite-backed InstalledTools", type: "output" as const },
  { text: "  • Uninstall command with --dry-run/--force/--purge", type: "output" as const },
  { text: "  • Multi-package-manager support (7 managers)", type: "output" as const },
];

const ChangelogPage = () => {
  const [expandedVersions, setExpandedVersions] = useState<Set<string>>(
    new Set(changelog.slice(0, 3).map((e) => e.version))
  );

  const toggle = (version: string) => {
    setExpandedVersions((prev) => {
      const next = new Set(prev);
      next.has(version) ? next.delete(version) : next.add(version);
      return next;
    });
  };

  const expandAll = () => setExpandedVersions(new Set(changelog.map((e) => e.version)));
  const collapseAll = () => setExpandedVersions(new Set());

  return (
    <DocsLayout>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-heading font-bold docs-h1">Changelog</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {changelog.length} releases tracked
          </p>
        </div>
        <div className="flex gap-2">
          <button onClick={expandAll} className="text-xs font-mono text-muted-foreground hover:text-foreground transition-colors px-2 py-1 rounded border border-border">
            Expand all
          </button>
          <button onClick={collapseAll} className="text-xs font-mono text-muted-foreground hover:text-foreground transition-colors px-2 py-1 rounded border border-border">
            Collapse all
          </button>
        </div>
      </div>

      <div className="mb-8">
        <TerminalDemo title="gitmap — version history" lines={terminalLines} autoPlay />
      </div>

      <div className="relative">
        {/* Timeline line */}
        <div className="absolute left-[15px] top-0 bottom-0 w-px bg-border" />

        <div className="space-y-2">
          {changelog.map((entry, i) => {
            const isOpen = expandedVersions.has(entry.version);
            const isLatest = i === 0;

            return (
              <div key={entry.version} className="relative pl-10">
                {/* Timeline dot */}
                <div className={`absolute left-[10px] top-3 h-[11px] w-[11px] rounded-full border-2 ${isLatest ? "border-primary bg-primary" : "border-muted-foreground/40 bg-background"}`} />

                <button
                  onClick={() => toggle(entry.version)}
                  className="w-full flex flex-col gap-1 px-4 py-2.5 rounded-lg border border-border bg-card hover:bg-muted/50 transition-colors text-left"
                >
                  <div className="flex items-center gap-3">
                    {isOpen ? (
                      <ChevronDown className="h-4 w-4 text-primary shrink-0" />
                    ) : (
                      <ChevronRight className="h-4 w-4 text-muted-foreground shrink-0" />
                    )}
                    <Tag className="h-3.5 w-3.5 text-primary shrink-0" />
                    <span className="font-mono font-semibold text-sm">{entry.version}</span>
                    {isLatest && (
                      <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-primary/10 text-foreground border border-primary/20 dark:bg-primary/15 dark:text-primary transition-colors duration-300 hover:border-primary/40 hover:shadow-sm hover:shadow-primary/10">
                        latest
                      </span>
                    )}
                    {entry.date && (
                      <span className="text-xs text-muted-foreground font-mono">
                        {entry.date}
                      </span>
                    )}
                    <span className="text-xs text-muted-foreground ml-auto">
                      {entry.items.length} change{entry.items.length !== 1 ? "s" : ""}
                    </span>
                  </div>
                  {entry.subtitle && (
                    <p className="text-xs text-muted-foreground/90 pl-[3.25rem] pr-4 leading-snug font-sans">
                      {entry.subtitle}
                    </p>
                  )}
                </button>

                <AnimatePresence initial={false}>
                  {isOpen && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: "auto", opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className="overflow-hidden"
                    >
                      <ul className="mt-1 ml-12 mr-2 space-y-1 pb-2 border-l border-border/60 pl-4">
                        {entry.items.map((item, j) => (
                          <li key={j} className="text-sm text-muted-foreground flex gap-2">
                            <span className="text-primary mt-1.5 shrink-0">•</span>
                            <span className="leading-relaxed">{item}</span>
                          </li>
                        ))}
                      </ul>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            );
          })}
        </div>
      </div>
    </DocsLayout>
  );
};

export default ChangelogPage;
