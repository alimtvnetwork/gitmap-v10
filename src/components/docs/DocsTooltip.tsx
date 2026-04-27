import {
  Children,
  cloneElement,
  isValidElement,
  ReactElement,
  ReactNode,
} from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  DOCS_TOOLTIP_ALIGN,
  DOCS_TOOLTIP_SIDE,
  DOCS_TOOLTIP_SIDE_OFFSET,
} from "@/components/docs/docsTooltip";

interface DocsTooltipProps {
  // The control the tooltip describes. Pass the trigger element
  // directly; this wrapper hands it to TooltipTrigger asChild so
  // the trigger keeps its own ref/keyboard semantics.
  children: ReactNode;
  // Tooltip body. Keep it short (one short phrase) — long text
  // belongs in inline help, not a hover tooltip.
  label: ReactNode;
  // Optional explicit accessible name. Defaults to `label` when it
  // is a string. Pass this when `label` is JSX (e.g. with an icon)
  // so screen-reader users still get a clean spoken name.
  ariaLabel?: string;
}

// Resolve the accessible name we want to expose on the trigger:
// 1. explicit `ariaLabel` prop (wins)
// 2. `label` when it is a plain string
// 3. otherwise undefined — the child must already supply its own.
const resolveAccessibleName = (
  label: ReactNode,
  ariaLabel: string | undefined,
): string | undefined => {
  if (ariaLabel) return ariaLabel;
  if (typeof label === "string") return label;
  return undefined;
};

// Inject aria-label onto the trigger child so keyboard / screen-reader
// users get the same information sighted users get from the tooltip.
// We never overwrite an aria-label the child already set — the child
// wins so callers can be more specific when they need to.
const withAccessibleName = (
  child: ReactNode,
  accessibleName: string | undefined,
): ReactNode => {
  if (!accessibleName) return child;
  if (!isValidElement(child)) return child;
  const childProps = child.props as { "aria-label"?: string };
  if (childProps["aria-label"]) return child;
  return cloneElement(child as ReactElement, { "aria-label": accessibleName });
};

// DocsTooltip is the ONLY way to attach a hover tooltip in the
// docs header / toolbars. Centralizing here means every tooltip
// shares the same side, offset, open/close delay (via the provider
// in App.tsx), AND the same keyboard/screen-reader contract:
// every icon-only trigger automatically receives an aria-label
// derived from `label` (or the explicit `ariaLabel` prop).
// Do NOT inline a raw <Tooltip> in docs surfaces.
export const DocsTooltip = ({ children, label, ariaLabel }: DocsTooltipProps) => {
  const accessibleName = resolveAccessibleName(label, ariaLabel);
  const trigger = withAccessibleName(children, accessibleName);
  return (
    <Tooltip>
      <TooltipTrigger asChild>{trigger}</TooltipTrigger>
      <TooltipContent
        side={DOCS_TOOLTIP_SIDE}
        sideOffset={DOCS_TOOLTIP_SIDE_OFFSET}
        align={DOCS_TOOLTIP_ALIGN}
      >
        {label}
      </TooltipContent>
    </Tooltip>
  );
};

export default DocsTooltip;
