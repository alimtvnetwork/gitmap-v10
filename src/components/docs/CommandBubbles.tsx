import { Link } from "react-router-dom";

interface Bubble {
  name: string;
  alias?: string;
  description: string;
}

const BUBBLES: Bubble[] = [
  { name: "scan", alias: "s", description: "Discover Git repos on disk" },
  { name: "clone", alias: "c", description: "Re-clone from a scan file" },
  { name: "clone-next", alias: "cn", description: "Clone next versioned iteration" },
  { name: "pull", alias: "p", description: "Pull latest for tracked repos" },
  { name: "watch", alias: "w", description: "Live status dashboard" },
  { name: "exec", alias: "x", description: "Run git across all repos" },
  { name: "release", alias: "r", description: "Branch, tag, push, attach" },
  { name: "as", alias: "s-alias", description: "Alias the current repo" },
  { name: "inject", alias: "inj", description: "Register folder + open VS Code" },
  { name: "cd", alias: "go", description: "Jump shell into a tracked repo" },
  { name: "group", alias: "g", description: "Manage repo groups" },
  { name: "changelog", alias: "cl", description: "View release notes" },
];

const CommandBubbles = () => {
  return (
    <section className="reveal py-8">
      <div className="mb-5 flex items-baseline justify-between gap-4">
        <h2 className="font-heading text-lg font-semibold text-foreground">
          Explore commands
        </h2>
        <Link
          to="/commands"
          className="rounded-sm text-xs font-sans text-muted-foreground hover:text-primary transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
        >
          View all →
        </Link>
      </div>

      <div className="flex flex-wrap gap-2">
        {BUBBLES.map((b) => (
          <Link
            key={b.name}
            to="/commands"
            title={b.description}
            aria-label={`${b.name} — ${b.description}`}
            className="btn-slide btn-slide-ghost group inline-flex items-center gap-2 rounded-full border border-border bg-card px-4 py-1.5 text-sm font-sans text-foreground hover:border-primary/50 hover:bg-secondary"
          >
            <code className="font-mono text-sm text-primary">{b.name}</code>
            {b.alias && (
              <span className="font-mono text-xs text-muted-foreground">
                {b.alias}
              </span>
            )}
          </Link>
        ))}
      </div>
    </section>
  );
};

export default CommandBubbles;
