import { cx } from "@/utils/cx";

export function Logo({ className }: { className?: string }) {
  return <img src="/logo.png" alt="" aria-hidden width={512} height={512} draggable={false} className={cx("size-8 shrink-0 select-none", className)} />;
}
