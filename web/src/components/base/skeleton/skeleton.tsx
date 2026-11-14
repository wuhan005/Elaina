"use client";

import type { CSSProperties, ReactNode } from "react";
import { cx, sortCx } from "@/utils/cx";

/**
 * Skeleton — placeholders in the shape of the content that is on its way.
 *
 *   <Skeleton shape="circle" className="size-10" />
 *   <SkeletonText lines={3} />
 *   <Skeleton isLoaded={user != null}><Avatar … /></Skeleton>
 *
 * Blocks are tinted from the text colour at 8%, so they show on any surface
 * in both themes (a fixed grey disappears on dark cards). A sheen of light
 * sweeps each one, drawn in layers like Spinner and Progress: a crisp 2px
 * line leads, a tight glow and a wide bloom trail it. It mirrors in RTL and
 * gives way to a gentle opacity pulse under reduced motion. The sheen lives
 * in globals.css (`.bui-skeleton`).
 *
 * With children, a Skeleton sizes itself to them: it lays the real content
 * out invisibly, then, once `isLoaded`, cross-fades the block away while the
 * content sharpens in from a 4px blur.
 *
 * Skeletons are hidden from assistive technology; mark the region that is
 * loading with `aria-busy`.
 */

export type SkeletonShape = "rect" | "text" | "circle";

const styles = sortCx({
  block: "bui-skeleton block bg-text-primary/8",
  shape: {
    rect: "rounded-xl",
    text: "rounded-md",
    circle: "rounded-full",
  },
});

const FADE = "transition-[opacity,filter] duration-300 ease-out motion-reduce:transition-none";

export interface SkeletonProps {
  /** `rect` (radius 12, default), `text` (radius 6), or `circle`. */
  shape?: SkeletonShape;
  /** With children: show them once true, blurring them in. */
  isLoaded?: boolean;
  /** Size and placement, e.g. `h-4 w-40` or `size-10`. */
  className?: string;
  style?: CSSProperties;
  /** Real content to measure the placeholder from and reveal when loaded. */
  children?: ReactNode;
}

export function Skeleton({ shape = "rect", isLoaded = false, className, style, children }: SkeletonProps) {
  if (children === undefined) {
    return <span aria-hidden className={cx(styles.block, "relative", styles.shape[shape], className)} style={style} />;
  }

  return (
    <span className={cx("relative block", className)} style={style}>
      <span
        aria-hidden={!isLoaded || undefined}
        inert={!isLoaded}
        className={cx("block", FADE, isLoaded ? "opacity-100 blur-[0px]" : "opacity-0 motion-safe:blur-[4px]")}
      >
        {children}
      </span>
      <span
        aria-hidden
        className={cx(styles.block, styles.shape[shape], "pointer-events-none absolute inset-0", FADE, isLoaded && "opacity-0")}
      />
    </span>
  );
}

export interface SkeletonTextProps {
  /** Number of lines. Default 3; the last one runs 60% wide. */
  lines?: number;
  /** Line height, matched to the text it stands in for. Default `h-3.5` (Body 1). */
  lineClassName?: string;
  className?: string;
}

export function SkeletonText({ lines = 3, lineClassName = "h-3.5", className }: SkeletonTextProps) {
  return (
    <span aria-hidden className={cx("flex w-full flex-col gap-2.5", className)}>
      {Array.from({ length: lines }, (_, index) => (
        <Skeleton
          key={index}
          shape="text"
          className={cx(lineClassName, index === lines - 1 && lines > 1 ? "w-3/5" : "w-full")}
        />
      ))}
    </span>
  );
}
