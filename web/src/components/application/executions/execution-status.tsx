import type { ExecutionSummary } from "@/api/api";
import { Chip, type ChipColor } from "@/components/base/badges/chip";

export type ExecutionStatus = ExecutionSummary["status"];

export const EXECUTION_STATUSES: { value: ExecutionStatus; label: string; color: ChipColor }[] = [
  { value: "success", label: "Success", color: "lime" },
  { value: "failed", label: "Failed", color: "rose" },
  { value: "timeout", label: "Timed out", color: "orange" },
  { value: "rejected", label: "Rejected", color: "yellow" },
  { value: "error", label: "Server error", color: "purple" },
];

const BY_VALUE = Object.fromEntries(EXECUTION_STATUSES.map((status) => [status.value, status]));

export function ExecutionStatusChip({ status }: { status: ExecutionStatus }) {
  const info = BY_VALUE[status];
  return (
    <Chip variant="caption" color={info?.color ?? "neutral"}>
      {info?.label ?? status}
    </Chip>
  );
}

export function formatDuration(nanoseconds: number): string {
  if (nanoseconds <= 0) return "—";
  const ms = nanoseconds / 1e6;
  if (ms < 1) return "<1 ms";
  if (ms < 1000) return `${Math.round(ms)} ms`;
  return `${(ms / 1000).toFixed(2)} s`;
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KiB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MiB`;
}
