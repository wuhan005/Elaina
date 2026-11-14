import { RiMagicLine } from "@remixicon/react";
import { cx } from "@/utils/cx";

export function ElainaMark({ size = "md", className }: { size?: "md" | "lg"; className?: string }) {
  return (
    <span
      aria-hidden
      className={cx(
        "flex shrink-0 items-center justify-center bg-linear-to-b from-accent-500 to-accent-600 shadow-nav-selected",
        size === "lg" ? "size-12 rounded-2xl" : "size-9 rounded-2lg",
        className,
      )}
    >
      <RiMagicLine className={cx("text-text-white", size === "lg" ? "size-6" : "size-5")} />
    </span>
  );
}
