"use client";

import { useEffect, useId, useState, useSyncExternalStore } from "react";
import type { ComponentType, ReactNode } from "react";
import {
  Notification,
  NotificationViewport,
  type NotificationAction,
  type NotificationAvatar,
  type NotificationPosition,
  type NotificationStatus,
} from "@/components/base/notification/notification";
import { cx } from "@/utils/cx";

/**
 * Toast — fire-and-forget notifications from anywhere in the app:
 *
 *   toast.success("Changes saved");
 *   toast.error("Upload failed", { description: "The file is over 20MB." });
 *   const id = toast("Exporting…", { duration: Infinity });
 *   toast.update(id, { title: "Export ready", status: "success", duration: 4000 });
 *   toast.dismiss(id);
 *
 * Render one <Toaster /> near the root of the app. It draws each toast as a
 * Notification card inside a NotificationViewport, so toasts share the
 * card, status icons, countdown bar, enter and exit motion, and the spring
 * that closes the stack's gaps. Extra Toasters render nothing, so a second
 * one in a nested layout never doubles every toast.
 *
 * The region is a polite live region (errors are announced as alerts). Every
 * countdown pauses while the pointer or keyboard focus is inside the stack,
 * so nobody loses a toast while reading it or reaching for its action.
 */

type IconComponent = ComponentType<{ className?: string; "aria-hidden"?: boolean | "true" | "false" }>;

export type ToastId = string | number;

export interface ToastOptions {
  /** Reuse an id to replace a toast that is still showing. */
  id?: ToastId;
  description?: ReactNode;
  status?: NotificationStatus;
  /** Custom leading icon; defaults to the status icon. */
  icon?: IconComponent;
  avatar?: NotificationAvatar;
  /** Small buttons under the message. Pressing one also dismisses the toast. */
  actions?: NotificationAction[];
  /** Milliseconds before it dismisses itself; `Infinity` keeps it until closed. Default from the Toaster (5000). */
  duration?: number;
  /** Show the close button. Default true. */
  dismissible?: boolean;
  /** Called once the toast has left the screen. */
  onDismiss?: () => void;
}

interface ToastRecord extends ToastOptions {
  id: ToastId;
  title: ReactNode;
  dismissed: boolean;
}

/* -------------------------------------------------------------------- store */

let toasts: ToastRecord[] = [];
let nextId = 1;
const listeners = new Set<() => void>();

function emit() {
  for (const listener of listeners) listener();
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function show(title: ReactNode, options: ToastOptions = {}): ToastId {
  const id = options.id ?? nextId++;
  const record: ToastRecord = { ...options, id, title, dismissed: false };
  const existing = toasts.findIndex((item) => item.id === id);
  toasts = existing >= 0 ? toasts.map((item, index) => (index === existing ? record : item)) : [...toasts, record];
  emit();
  return id;
}

function remove(id: ToastId) {
  const record = toasts.find((item) => item.id === id);
  toasts = toasts.filter((item) => item.id !== id);
  emit();
  record?.onDismiss?.();
}

/** Show a toast. Returns its id, for `toast.update` and `toast.dismiss`. */
export function toast(title: ReactNode, options?: ToastOptions): ToastId {
  return show(title, options);
}

toast.success = (title: ReactNode, options?: Omit<ToastOptions, "status">) => show(title, { ...options, status: "success" });
toast.error = (title: ReactNode, options?: Omit<ToastOptions, "status">) => show(title, { ...options, status: "error" });
toast.info = (title: ReactNode, options?: Omit<ToastOptions, "status">) => show(title, { ...options, status: "information" });

/** Change a toast that is showing. A new `duration` restarts its countdown. */
toast.update = (id: ToastId, options: ToastOptions & { title?: ReactNode }) => {
  const current = toasts.find((item) => item.id === id && !item.dismissed);
  if (!current) return;
  const { title = current.title, ...rest } = options;
  show(title, { ...current, ...rest, id });
};

/** Dismiss one toast, or every toast when called without an id. */
toast.dismiss = (id?: ToastId) => {
  toasts = toasts.map((item) => (id === undefined || item.id === id ? { ...item, dismissed: true } : item));
  emit();
};

/* ------------------------------------------------------------------ toaster */

// Only the first mounted Toaster draws the stack; the rest stand by and take
// over if it unmounts.
let owners: string[] = [];
const ownerListeners = new Set<() => void>();

function subscribeOwners(listener: () => void) {
  ownerListeners.add(listener);
  return () => {
    ownerListeners.delete(listener);
  };
}

export interface ToasterProps {
  /** Edge or corner the stack is anchored to. Default `bottom-end`. */
  position?: NotificationPosition;
  /** How many toasts show at once; older ones leave as new ones arrive. Default 3. */
  limit?: number;
  /** Default countdown in milliseconds. Default 5000. */
  duration?: number;
  /** Accessible name of the region. */
  "aria-label"?: string;
  /** Accessible name of each toast's close button. */
  closeLabel?: string;
  className?: string;
}

export function Toaster({
  position = "bottom-end",
  limit = 3,
  duration = 5000,
  "aria-label": ariaLabel = "Notifications",
  closeLabel,
  className,
}: ToasterProps) {
  const ownerId = useId();
  const items = useSyncExternalStore(subscribe, () => toasts, () => toasts);
  const isOwner = useSyncExternalStore(subscribeOwners, () => owners[0] === ownerId, () => false);
  const [hovered, setHovered] = useState(false);
  const [focused, setFocused] = useState(false);

  useEffect(() => {
    owners = [...owners, ownerId];
    ownerListeners.forEach((listener) => listener());
    return () => {
      owners = owners.filter((owner) => owner !== ownerId);
      ownerListeners.forEach((listener) => listener());
    };
  }, [ownerId]);

  // Past the limit, the oldest toasts play their exit and make room.
  useEffect(() => {
    if (!isOwner) return;
    const active = items.filter((item) => !item.dismissed);
    if (active.length > limit) {
      const overflow = new Set(active.slice(0, active.length - limit).map((item) => item.id));
      toasts = toasts.map((item) => (overflow.has(item.id) ? { ...item, dismissed: true } : item));
      emit();
    }
  }, [items, limit, isOwner]);

  // A toast removed from under the pointer or focus fires no leave or blur,
  // so an emptied stack resets both, or the next toast would start paused.
  const empty = items.length === 0;
  const [wasEmpty, setWasEmpty] = useState(empty);
  if (empty !== wasEmpty) {
    setWasEmpty(empty);
    if (empty) {
      setHovered(false);
      setFocused(false);
    }
  }

  if (!isOwner) return null;

  return (
    <NotificationViewport
      position={position}
      aria-label={ariaLabel}
      role="region"
      aria-live="polite"
      aria-relevant="additions text"
      className={cx("z-130", className)}
      onPointerEnter={() => setHovered(true)}
      onPointerLeave={() => setHovered(false)}
      onFocus={() => setFocused(true)}
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget as Node | null)) setFocused(false);
      }}
    >
      {items.map((item) => {
        const timeout = item.duration ?? duration;
        return (
          <Notification
            // A toast replaced under the same id remounts, so its countdown restarts.
            key={`${item.id}:${timeout}`}
            title={item.title}
            description={item.description}
            status={item.status}
            icon={item.icon}
            avatar={item.avatar}
            dismissible={item.dismissible ?? true}
            closeLabel={closeLabel}
            actions={item.actions?.map((action) => ({
              ...action,
              onClick: () => {
                action.onClick?.();
                toast.dismiss(item.id);
              },
            }))}
            autoDismissDuration={Number.isFinite(timeout) && timeout > 0 ? timeout : undefined}
            paused={hovered || focused}
            dismissed={item.dismissed}
            introDelay={0}
            onDismiss={() => remove(item.id)}
          />
        );
      })}
    </NotificationViewport>
  );
}
