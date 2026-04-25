import { useState } from "react";
import { Keyboard, ChevronDown } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

/**
 * TabOrderMap — collapsible accessibility helper that visualizes the
 * keyboard tab sequence across the homepage hero card, hero CTA buttons,
 * and command bubble grid. Numbers reflect natural DOM order.
 *
 * Purely presentational. The real focus order is driven by the live DOM —
 * this diagram is a static reference that mirrors it.
 */

interface FocusNode {
  step: number;
  label: string;
  hint?: string;
}

// Hero card surfaces (Install + Uninstall tab strips share this pattern)
const HERO_CARD: FocusNode[] = [
  { step: 1, label: "Install · OS tabs", hint: "Windows / Linux-macOS" },
  { step: 2, label: "Install · Copy", hint: "Copy install command" },
  { step: 3, label: "Uninstall · OS tabs", hint: "Windows / Linux-macOS" },
  { step: 4, label: "Uninstall · Copy", hint: "Copy uninstall command" },
];

const HERO_BUTTONS: FocusNode[] = [
  { step: 5, label: "Get Started", hint: "Primary CTA" },
  { step: 6, label: "View Commands", hint: "Secondary CTA" },
];

const BUBBLE_HEADER: FocusNode[] = [
  { step: 7, label: "View all →", hint: "Skip to /commands" },
];

const BUBBLES = [
  "scan", "clone", "clone-next", "pull",
  "watch", "exec", "release", "as",
  "inject", "cd", "group", "changelog",
];

const Badge = ({ n }: { n: number }) => (
  <span
    aria-hidden="true"
    className="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-primary text-[10px] font-mono font-bold text-primary-foreground shadow-sm"
  >
    {n}
  </span>
);

const Row = ({ node }: { node: FocusNode }) => (
  <div className="flex items-center gap-3 rounded-md border border-border bg-card/60 px-3 py-2">
    <Badge n={node.step} />
    <div className="min-w-0 flex-1">
      <div className="font-sans text-sm text-foreground">{node.label}</div>
      {node.hint && (
        <div className="font-mono text-[11px] text-muted-foreground">{node.hint}</div>
      )}
    </div>
  </div>
);

const TabOrderMap = () => {
  const [open, setOpen] = useState(false);
  const bubbleStart = BUBBLE_HEADER.length + HERO_BUTTONS.length + HERO_CARD.length + 1;

  return (
    <section className="py-6">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
        aria-controls="tab-order-map-panel"
        className="inline-flex items-center gap-2 rounded-md border border-border bg-card px-3 py-1.5 text-xs font-sans text-muted-foreground hover:text-foreground hover:border-primary/40 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      >
        <Keyboard className="h-3.5 w-3.5" />
        <span>{open ? "Hide" : "Show"} tab order map</span>
        <motion.span
          animate={{ rotate: open ? 180 : 0 }}
          transition={{ duration: 0.2 }}
          className="inline-flex"
        >
          <ChevronDown className="h-3.5 w-3.5" />
        </motion.span>
      </button>

      <AnimatePresence initial={false}>
        {open && (
          <motion.div
            id="tab-order-map-panel"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.25, ease: "easeInOut" }}
            className="overflow-hidden"
          >
            <div className="mt-4 rounded-lg border border-border bg-muted/20 p-5">
              <p className="mb-4 font-sans text-xs text-muted-foreground">
                Keyboard focus moves in this order. Use{" "}
                <kbd className="rounded border border-border bg-card px-1.5 py-0.5 font-mono text-[10px]">Tab</kbd>{" "}
                to advance and{" "}
                <kbd className="rounded border border-border bg-card px-1.5 py-0.5 font-mono text-[10px]">Shift+Tab</kbd>{" "}
                to go back.
              </p>

              <div className="grid gap-5 md:grid-cols-2">
                {/* Wireframe: hero card + buttons */}
                <div className="space-y-4">
                  <div className="rounded-lg border border-dashed border-border bg-card/30 p-3">
                    <div className="mb-2 font-heading text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                      Hero terminal card
                    </div>
                    <div className="space-y-1.5">
                      {HERO_CARD.map((n) => <Row key={n.step} node={n} />)}
                    </div>
                  </div>

                  <div className="rounded-lg border border-dashed border-border bg-card/30 p-3">
                    <div className="mb-2 font-heading text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                      Hero CTAs
                    </div>
                    <div className="space-y-1.5">
                      {HERO_BUTTONS.map((n) => <Row key={n.step} node={n} />)}
                    </div>
                  </div>
                </div>

                {/* Wireframe: bubble grid */}
                <div className="rounded-lg border border-dashed border-border bg-card/30 p-3">
                  <div className="mb-2 flex items-center justify-between">
                    <span className="font-heading text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                      Command bubbles
                    </span>
                    <Row node={BUBBLE_HEADER[0]} />
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {BUBBLES.map((name, i) => (
                      <span
                        key={name}
                        className="inline-flex items-center gap-1.5 rounded-full border border-border bg-card px-2.5 py-1 font-mono text-xs text-foreground"
                      >
                        <Badge n={bubbleStart + i} />
                        {name}
                      </span>
                    ))}
                  </div>
                </div>
              </div>

              <p className="mt-4 font-sans text-[11px] text-muted-foreground">
                Tip: every focusable element has a 3px ring outline so the active
                step is always visible — no hover required.
              </p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </section>
  );
};

export default TabOrderMap;
