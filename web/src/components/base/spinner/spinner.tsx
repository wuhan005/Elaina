"use client";

import type { CSSProperties } from "react";
import { ProgressBar as AriaProgressBar } from "react-aria-components";
import { cx } from "@/utils/cx";

/**
 * Spinner — an orbiting light for work of unknown length.
 *
 *   <Spinner />                                  // 20px, accent
 *   <Spinner size="sm" className="text-white" /> // inside a primary button
 *
 * The light is drawn the way Composer Loader draws its band: the same arc
 * at three scales, a crisp line with a rounded head, a tight glow and a wide
 * bloom, each on its own blur. The soft layers carry shorter arcs and turn a
 * few degrees behind, so the sharp head always leads and the glow gathers
 * just behind it like the tail of a comet. A faint track completes the
 * circle. Everything is drawn in `currentColor`, so a text colour sets it.
 *
 * Announced as an indeterminate progress bar labelled "Loading" (override
 * with `aria-label`). Pass `aria-hidden` where the surrounding control
 * already says it is busy, e.g. a pending button. Reduced motion slows the
 * orbit rather than stopping it: the motion is the message.
 */

export type SpinnerSize = "xs" | "sm" | "md" | "lg" | "xl";

/** Box size and line width in px. */
const SIZES: Record<SpinnerSize, { box: number; line: number }> = {
  xs: { box: 12, line: 1.5 },
  sm: { box: 16, line: 2 },
  md: { box: 20, line: 2 },
  lg: { box: 24, line: 2.5 },
  xl: { box: 32, line: 3 },
};

/** One scale of the light: a conic arc masked to a ring, blurred by its wrapper. */
function Arc({
  box,
  line,
  width,
  blur,
  opacity,
  arc,
  lag,
}: {
  box: number;
  line: number;
  /** Ring thickness, px. */
  width: number;
  blur: number;
  opacity: number;
  /** Arc length in degrees; the head sits at 12 o'clock before the turn. */
  arc: number;
  /** Degrees this layer trails the crisp head. */
  lag: number;
}) {
  const center = box / 2 - line / 2;
  const inner = center - width / 2;
  const outer = center + width / 2;
  // The wrapper is grown to hold the ring plus its blur, so nothing clips.
  const bleed = Math.max(0, outer + blur * 2 - box / 2);
  const ring = `radial-gradient(circle, transparent ${inner}px, #000 ${inner + 0.5}px, #000 ${outer - 0.5}px, transparent ${outer}px)`;
  const style: CSSProperties = {
    inset: -bleed,
    filter: blur > 0 ? `blur(${blur}px)` : undefined,
    opacity,
    rotate: `${-lag}deg`,
  };
  return (
    <span aria-hidden className="absolute" style={style}>
      <span
        className="absolute inset-0"
        style={{
          background: `conic-gradient(transparent ${360 - arc}deg, currentColor 360deg)`,
          maskImage: ring,
          WebkitMaskImage: ring,
        }}
      />
    </span>
  );
}

export interface SpinnerProps {
  /** 12, 16, 20 (default), 24, or 32px. */
  size?: SpinnerSize;
  /** Draw the faint full circle under the light. Default true. */
  track?: boolean;
  /** Accessible name. Default "Loading". */
  "aria-label"?: string;
  /** Hide it from assistive technology when a parent already announces the wait. */
  "aria-hidden"?: boolean;
  /** Colour comes from the text colour, e.g. `text-white`. Default accent. */
  className?: string;
}

export function SpinnerGlyph({ size = "md", track = true, className }: Omit<SpinnerProps, "aria-label" | "aria-hidden">) {
  const { box, line } = SIZES[size];
  return (
    <span
      aria-hidden
      className={cx("relative inline-flex shrink-0 text-accent-500", className)}
      style={{ width: box, height: box }}
    >
      {track ? (
        <span className="absolute inset-0 rounded-full border-current opacity-15" style={{ borderWidth: line }} />
      ) : null}
      <span className="absolute inset-0 animate-spin [animation-duration:900ms] motion-reduce:[animation-duration:2400ms]">
        <Arc box={box} line={line} width={line * 4} blur={line * 1.6} opacity={0.32} arc={110} lag={22} />
        <Arc box={box} line={line} width={line * 2} blur={line * 0.6} opacity={0.6} arc={200} lag={9} />
        <Arc box={box} line={line} width={line} blur={0} opacity={1} arc={290} lag={0} />
        {/* The crisp line's rounded head, at the arc's leading end. */}
        <span
          className="absolute left-1/2 top-0 rounded-full bg-current"
          style={{ width: line, height: line, translate: "-50% 0" }}
        />
      </span>
    </span>
  );
}

export function Spinner({ size = "md", track = true, "aria-label": ariaLabel = "Loading", "aria-hidden": ariaHidden, className }: SpinnerProps) {
  if (ariaHidden) return <SpinnerGlyph size={size} track={track} className={className} />;
  return (
    <AriaProgressBar isIndeterminate aria-label={ariaLabel} className="inline-flex shrink-0">
      <SpinnerGlyph size={size} track={track} className={className} />
    </AriaProgressBar>
  );
}
