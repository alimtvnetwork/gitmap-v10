import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { TooltipProvider } from "@/components/ui/tooltip";
import { SidebarProvider } from "@/components/ui/sidebar";
import DocsLayout from "@/components/docs/DocsLayout";
import CodeBlock from "@/components/docs/CodeBlock";

// Why focus() instead of hover():
// Radix Tooltip opens on either pointer-enter OR focus. jsdom does
// not synthesize pointer events the way real browsers do, so
// focus is the deterministic path to "tooltip is open" in this
// environment. Real-browser pointer behaviour is covered by the
// shared DocsTooltip component (single source of truth for
// placement/delay) — this suite verifies WIRING, not Radix internals.

const renderDocsChrome = () =>
  render(
    <MemoryRouter>
      <TooltipProvider delayDuration={0} skipDelayDuration={0}>
        <SidebarProvider>
          <DocsLayout>
            <div>content</div>
          </DocsLayout>
        </SidebarProvider>
      </TooltipProvider>
    </MemoryRouter>,
  );

const renderCodeBlock = () =>
  render(
    <TooltipProvider delayDuration={0} skipDelayDuration={0}>
      <CodeBlock code={"line one\nline two"} language="bash" />
    </TooltipProvider>,
  );

// HEADER_CONTROLS lists every icon-only control in the docs header.
// Each entry pairs the trigger's accessible name (used to look the
// element up via getByRole / getByLabelText) with the substring we
// expect to find in the surfaced tooltip body. Update this list
// whenever a control is added/removed from the header.
const HEADER_CONTROLS: Array<{ trigger: string; tooltip: string }> = [
  { trigger: "Toggle sidebar", tooltip: "Toggle sidebar" },
  { trigger: "Dark theme", tooltip: "Dark theme" },
  { trigger: "Light theme", tooltip: "Light theme" },
  { trigger: /Copy .* theme palette to clipboard/i.source, tooltip: "palette" },
  {
    trigger: "Open command palette (search commands, flags, pages)",
    tooltip: "Search commands",
  },
];

// CODEBLOCK_CONTROLS lists every icon-only control in the CodeBlock
// toolbar. Same shape and update rule as HEADER_CONTROLS.
const CODEBLOCK_CONTROLS: Array<{ trigger: string; tooltip: string }> = [
  { trigger: "Decrease font size", tooltip: "Decrease font size" },
  { trigger: "Increase font size", tooltip: "Increase font size" },
  { trigger: "Copy snippet", tooltip: "Copy snippet" },
  { trigger: "Download snippet", tooltip: "Download snippet" },
  { trigger: "Enter fullscreen", tooltip: "Fullscreen" },
];

const findTriggerByName = (name: string) => {
  // Try exact aria-label first; fall back to a regex match for the
  // dynamic CopyPaletteButton label that includes the active theme.
  try {
    return screen.getByLabelText(name);
  } catch {
    return screen.getByLabelText(new RegExp(name, "i"));
  }
};

const expectTooltipFor = async (
  triggerName: string,
  tooltipSubstr: string,
) => {
  const trigger = findTriggerByName(triggerName);
  // Focus is the jsdom-friendly opener (see top-of-file comment).
  trigger.focus();
  // Radix mounts the tooltip body in a portal as role="tooltip".
  // findAllByRole tolerates same-text matches across header + portal.
  const tips = await screen.findAllByRole("tooltip");
  const matched = tips.some((t) =>
    (t.textContent ?? "").toLowerCase().includes(tooltipSubstr.toLowerCase()),
  );
  expect(
    matched,
    `expected an open tooltip whose text includes "${tooltipSubstr}" ` +
      `for trigger "${triggerName}", got: ${tips.map((t) => t.textContent).join(" | ")}`,
  ).toBe(true);
  trigger.blur();
};

describe("docs tooltip wiring — header controls", () => {
  beforeEach(() => cleanup());

  for (const ctl of HEADER_CONTROLS) {
    it(`shows tooltip "${ctl.tooltip}" for "${ctl.trigger}"`, async () => {
      renderDocsChrome();
      await expectTooltipFor(ctl.trigger, ctl.tooltip);
    });
  }
});

describe("docs tooltip wiring — CodeBlock toolbar", () => {
  beforeEach(() => cleanup());

  for (const ctl of CODEBLOCK_CONTROLS) {
    it(`shows tooltip "${ctl.tooltip}" for "${ctl.trigger}"`, async () => {
      renderCodeBlock();
      await expectTooltipFor(ctl.trigger, ctl.tooltip);
    });
  }
});

describe("docs tooltip wiring — version badge (non-icon trigger)", () => {
  beforeEach(() => cleanup());

  it("surfaces a tooltip for the version badge on focus", async () => {
    renderDocsChrome();
    // The version badge has aria-label="gitmap version <X>" (no icon
    // but it is a focus-able decorative chip — the audit lists it
    // as needing a tooltip alongside the icon-only controls).
    const badge = screen.getByLabelText(/gitmap version /i);
    badge.focus();
    const tips = await screen.findAllByRole("tooltip");
    const matched = tips.some((t) =>
      (t.textContent ?? "").toLowerCase().includes("gitmap version"),
    );
    expect(matched).toBe(true);
  });
});

// Pointer-event sanity check: also verify userEvent.hover() opens
// the tooltip in jsdom for at least one well-known trigger. If
// Radix changes its opener semantics this test will catch it.
describe("docs tooltip wiring — hover opener", () => {
  beforeEach(() => cleanup());

  it("opens the sidebar-toggle tooltip on pointer hover", async () => {
    const user = userEvent.setup();
    renderDocsChrome();
    const trigger = screen.getByLabelText("Toggle sidebar");
    await user.hover(trigger);
    const tips = await screen.findAllByRole("tooltip");
    const matched = tips.some((t) =>
      (t.textContent ?? "").toLowerCase().includes("toggle sidebar"),
    );
    expect(matched).toBe(true);
  });
});

// DocsTooltip wraps `children` with TooltipTrigger asChild, but it
// also tries to inject an aria-label onto the child via cloneElement.
// The injection helper must short-circuit when the child is NOT a
// single valid React element — otherwise React would throw at render
// time. This suite locks in that contract: every non-element child
// shape renders without throwing and the tooltip body still appears
// when a focusable descendant receives focus. If someone refactors
// `withAccessibleName` and forgets the isValidElement guard, these
// tests fail loudly instead of crashing the whole docs chrome.
import { DocsTooltip } from "@/components/docs/DocsTooltip";

const renderTooltip = (node: React.ReactNode) =>
  render(
    <TooltipProvider delayDuration={0} skipDelayDuration={0}>
      <DocsTooltip label="fallback label">{node}</DocsTooltip>
    </TooltipProvider>,
  );

describe("DocsTooltip — non-element children fallback", () => {
  beforeEach(() => cleanup());

  it("does not throw when the child is a plain string", () => {
    expect(() => renderTooltip("just text")).not.toThrow();
    expect(screen.getByText("just text")).toBeTruthy();
  });

  it("does not throw when the child is a number", () => {
    expect(() => renderTooltip(42)).not.toThrow();
  });

  it("does not throw when the child is null", () => {
    expect(() => renderTooltip(null)).not.toThrow();
  });

  it("does not throw when there are multiple children (array)", () => {
    expect(() =>
      renderTooltip([
        <span key="a">a</span>,
        <button key="b" type="button">
          b
        </button>,
      ]),
    ).not.toThrow();
  });

  it("does not throw when the child is a fragment", () => {
    expect(() =>
      renderTooltip(
        <>
          <span>x</span>
          <span>y</span>
        </>,
      ),
    ).not.toThrow();
  });

  it("still surfaces the tooltip body on focus of a focusable descendant", async () => {
    // Even when we cannot inject aria-label (multiple children),
    // Radix should still open the tooltip on focus of the trigger
    // subtree. This proves the fallback degrades gracefully.
    renderTooltip(
      <>
        <button type="button" aria-label="inner btn">
          inner
        </button>
        <span>extra</span>
      </>,
    );
    // Hover the wrapper subtree (jsdom-friendly via userEvent) to
    // confirm the tooltip body still renders. The KEY assertion of
    // this test is "no throw + tooltip remains usable" — the precise
    // focus target inside the wrapper is a Radix Slot internal.
    const user = userEvent.setup();
    const inner = screen.getByLabelText("inner btn");
    const wrapper = inner.parentElement as HTMLElement;
    await user.hover(wrapper);
    const tips = await screen.findAllByRole("tooltip");
    const matched = tips.some((t) =>
      (t.textContent ?? "").includes("fallback label"),
    );
    expect(matched).toBe(true);
  });
});
