"use client";

import {
  Children,
  createContext,
  isValidElement,
  useContext,
  useId,
  useLayoutEffect,
  useState,
} from "react";
import type { ComponentType, ReactElement, ReactNode } from "react";
import {
  Dialog as AriaDialog,
  DialogTrigger as AriaDialogTrigger,
  Heading as AriaHeading,
  Modal as AriaModal,
  ModalOverlay as AriaModalOverlay,
  OverlayTriggerStateContext,
  Pressable,
} from "react-aria-components";
import type { DialogRenderProps, DialogTriggerProps, ModalOverlayProps } from "react-aria-components";
import { RiErrorWarningLine } from "@remixicon/react";
import { Button } from "@/components/base/buttons/button";
import { CloseButton } from "@/components/base/buttons/close-button";
import { SpinnerGlyph } from "@/components/base/spinner/spinner";
import { useDirection } from "@/components/foundations/direction/direction";
import { ScrollFade, useScrollFade } from "@/components/foundations/scroll-fade/scroll-fade";
import { cx, sortCx } from "@/utils/cx";

/**
 * Modal — a centred dialog over a dimmed page, built on React Aria's
 * ModalOverlay / Modal / Dialog: focus is trapped inside and returns to the
 * trigger on close, the page behind stops scrolling, and Escape or a press on
 * the backdrop dismisses it.
 *
 *   <ModalTrigger>
 *     <Button>Invite teammates</Button>
 *     <Modal>
 *       {({ close }) => (
 *         <>
 *           <ModalHeader title="Invite teammates" description="…" />
 *           <ModalBody>…</ModalBody>
 *           <ModalFooter>
 *             <Button variant="secondary" onClick={close}>Cancel</Button>
 *             <Button onClick={close}>Send invites</Button>
 *           </ModalFooter>
 *         </>
 *       )}
 *     </Modal>
 *   </ModalTrigger>
 *
 * Or control it from your own state with `isOpen` / `onOpenChange`, no trigger.
 * `AlertDialog` is the destructive-confirm preset: role="alertdialog", no
 * backdrop dismissal, focus starts on Cancel.
 *
 * Surface and motion follow Settings Modal: black/70 backdrop fading over
 * 300ms; the panel fades and sharpens from scale 0.85 and a 4px blur over
 * 300ms on the iOS sheet curve, and plays the same transition in reverse
 * before it unmounts. Reduced motion keeps the fade and drops the scale and
 * blur.
 */

type IconComponent = ComponentType<{ className?: string; "aria-hidden"?: boolean | "true" | "false" }>;

export type ModalSize = "sm" | "md" | "lg" | "xl";

const styles = sortCx({
  overlay: "group/modal fixed inset-x-0 top-0 z-100 flex h-(--visual-viewport-height) items-center justify-center p-4 sm:p-6",
  backdrop: [
    "absolute inset-0 bg-black/70",
    "transition-opacity duration-300 ease-out",
    "group-data-[entering]/modal:opacity-0 group-data-[exiting]/modal:opacity-0",
  ].join(" "),
  modal: [
    "relative flex max-h-full w-full flex-col outline-none",
    "transform-gpu transition-[opacity,scale,filter] duration-300 ease-[cubic-bezier(0.32,0.72,0,1)]",
    "data-[entering]:opacity-0 data-[exiting]:opacity-0",
    "motion-safe:data-[entering]:scale-[0.85] motion-safe:data-[entering]:blur-[4px]",
    "motion-safe:data-[exiting]:scale-[0.85] motion-safe:data-[exiting]:blur-[4px]",
  ].join(" "),
  dialog: [
    // overflow-CLIP, not hidden: a hidden box is still programmatically
    // scrollable, so focusing a field clipped by the body's scroller could
    // scroll the whole panel instead (see Settings Modal).
    "flex min-h-0 flex-col gap-4 overflow-clip py-6 outline-none",
    // Settings Modal's surface: page white in light mode, and in dark mode a
    // ground darker than the inputs and menus it holds, so they keep an edge.
    "rounded-3xl border border-separator-border bg-background-full shadow-xs",
  ].join(" "),
  size: {
    sm: "max-w-[400px]",
    md: "max-w-[480px]",
    lg: "max-w-[640px]",
    xl: "max-w-[800px]",
  },
  icon: {
    default: "bg-background-tertiary-default text-foreground-icon-secondary",
    danger: "bg-notification-error-background text-notification-error-foreground",
  },
});

/* ------------------------------------------------------------------ context */

interface ModalContextValue {
  descriptionId: string;
  setDescribed: (described: boolean) => void;
}

const ModalContext = createContext<ModalContextValue | null>(null);

/** Closes the surrounding Modal, Drawer, or AlertDialog. */
export function useModalClose() {
  const state = useContext(OverlayTriggerStateContext);
  return () => state?.close();
}

/* ------------------------------------------------------------------ trigger */

export interface ModalTriggerProps extends Omit<DialogTriggerProps, "children"> {
  /**
   * The trigger, then the Modal (or AlertDialog). The trigger can be any
   * BoardUI button: anything that forwards its ref and spreads its props.
   * React Aria's own Button already speaks to DialogTrigger; use that
   * directly instead of wrapping it here.
   */
  children: ReactNode;
}

export function ModalTrigger({ children, ...props }: ModalTriggerProps) {
  const [trigger, ...rest] = Children.toArray(children);
  return (
    <AriaDialogTrigger {...props}>
      {isValidElement(trigger) ? <Pressable>{trigger as ReactElement<Record<string, unknown>, string>}</Pressable> : trigger}
      {rest}
    </AriaDialogTrigger>
  );
}

/* -------------------------------------------------------------------- modal */

export interface ModalProps extends Omit<ModalOverlayProps, "children" | "className" | "style"> {
  /** Panel width: 400, 480 (default), 640, or 800px. */
  size?: ModalSize;
  /** `alertdialog` for confirmations that interrupt the user; see AlertDialog. */
  role?: "dialog" | "alertdialog";
  /** Accessible name when the content has no ModalHeader title. */
  "aria-label"?: string;
  /** Classes on the panel surface (the dialog). */
  className?: string;
  /** Classes on the full-screen overlay, e.g. to align the panel to the top. */
  overlayClassName?: string;
  children: ReactNode | ((renderProps: DialogRenderProps) => ReactNode);
}

export function Modal({
  size = "md",
  role,
  "aria-label": ariaLabel,
  className,
  overlayClassName,
  isDismissable = true,
  children,
  ...props
}: ModalProps) {
  const direction = useDirection();

  return (
    <AriaModalOverlay {...props} dir={direction} isDismissable={isDismissable} className={cx(styles.overlay, overlayClassName)}>
      <div aria-hidden className={styles.backdrop} />
      <AriaModal className={cx(styles.modal, styles.size[size])}>
        <ModalDialog role={role} aria-label={ariaLabel} className={className}>
          {children}
        </ModalDialog>
      </AriaModal>
    </AriaModalOverlay>
  );
}

export interface ModalDialogProps {
  role?: "dialog" | "alertdialog";
  "aria-label"?: string;
  /** Classes merged onto the panel surface. */
  className?: string;
  children: ReactNode | ((renderProps: DialogRenderProps) => ReactNode);
}

/**
 * The panel surface and dialog element shared by Modal and Drawer: the
 * ModalHeader title names it and the description is wired to
 * aria-describedby. Render it inside a React Aria Modal.
 */
export function ModalDialog({ role, "aria-label": ariaLabel, className, children }: ModalDialogProps) {
  const descriptionId = useId();
  const [described, setDescribed] = useState(false);

  return (
    <AriaDialog
      role={role}
      aria-label={ariaLabel}
      aria-describedby={described ? descriptionId : undefined}
      className={cx(styles.dialog, className)}
    >
      {(renderProps) => (
        <ModalContext.Provider value={{ descriptionId, setDescribed }}>
          {typeof children === "function" ? children(renderProps) : children}
        </ModalContext.Provider>
      )}
    </AriaDialog>
  );
}

/* ------------------------------------------------------------------ content */

export interface ModalHeaderProps {
  title: ReactNode;
  /** Supporting copy under the title; announced as the dialog's description. */
  description?: ReactNode;
  /** Remix Icon shown in a 40px circle above the title. */
  icon?: IconComponent;
  /** `danger` tints the icon circle for destructive dialogs. */
  tone?: "default" | "danger";
  /** Show the dismiss button in the top corner. Default true. */
  showCloseButton?: boolean;
  closeLabel?: string;
  className?: string;
  /** Extra content under the description, e.g. tabs or a step indicator. */
  children?: ReactNode;
}

export function ModalHeader({
  title,
  description,
  icon: Icon,
  tone = "default",
  showCloseButton = true,
  closeLabel = "Close",
  className,
  children,
}: ModalHeaderProps) {
  const context = useContext(ModalContext);
  const close = useModalClose();
  const setDescribed = context?.setDescribed;
  const hasDescription = description != null && description !== false;

  useLayoutEffect(() => {
    if (!hasDescription || !setDescribed) return;
    setDescribed(true);
    return () => setDescribed(false);
  }, [hasDescription, setDescribed]);

  return (
    <div className={cx("flex shrink-0 items-start gap-3 px-6", className)}>
      <div className="flex min-w-0 flex-1 flex-col">
        {Icon ? (
          <span className={cx("mb-4 flex size-10 shrink-0 items-center justify-center rounded-full", styles.icon[tone])}>
            <Icon className="size-5" aria-hidden />
          </span>
        ) : null}
        <AriaHeading slot="title" className="text-title-3-medium text-text-primary">
          {title}
        </AriaHeading>
        {hasDescription ? (
          <p id={context?.descriptionId} className="mt-1 text-body-regular text-text-secondary">
            {description}
          </p>
        ) : null}
        {children}
      </div>
      {showCloseButton ? (
        <CloseButton size="sm" aria-label={closeLabel} onClick={close} className="-me-1 shrink-0" />
      ) : null}
    </div>
  );
}

export interface ModalBodyProps {
  /** Classes on the scrolling region (it carries px-6 by default). */
  className?: string;
  children: ReactNode;
}

/**
 * The content between the header and the footer. It scrolls on its own when
 * the panel reaches the viewport's height, with the header and footer held
 * in place and a progressive blur softening the top edge once scrolled.
 */
export function ModalBody({ className, children }: ModalBodyProps) {
  const fade = useScrollFade();
  return (
    // flex-1 fills a full-height panel (Drawer) and is inert in an
    // auto-height one: a 0% basis against an indefinite height acts as auto.
    <div className="relative flex min-h-0 flex-1 flex-col">
      <div
        onScroll={fade.onScroll}
        className={cx("min-h-0 flex-1 overflow-y-auto overscroll-contain px-6 py-1 text-body-regular text-text-secondary", className)}
      >
        {children}
      </div>
      <ScrollFade progress={fade.progress} surfaceClassName="from-background-full" />
    </div>
  );
}

export function ModalFooter({ className, children }: { className?: string; children: ReactNode }) {
  return (
    <div
      className={cx(
        // Stacked full-width on phones (primary action on top), a trailing
        // row from sm up. 8px extra over the dialog's gap-4 sets the actions
        // apart from the content.
        "mt-2 flex shrink-0 flex-col-reverse gap-2 px-6 sm:flex-row sm:items-center sm:justify-end",
        className,
      )}
    >
      {children}
    </div>
  );
}

/* ------------------------------------------------------------- alert dialog */

export interface AlertDialogProps
  extends Omit<ModalProps, "children" | "role" | "aria-label" | "isDismissable"> {
  title: ReactNode;
  description?: ReactNode;
  /** `danger` (default) pairs a tinted warning icon with a danger confirm button. */
  variant?: "danger" | "default";
  /** Leading icon; `false` hides it. Defaults to a warning glyph. */
  icon?: IconComponent | false;
  confirmLabel?: ReactNode;
  cancelLabel?: ReactNode;
  /**
   * Runs on confirm. Return a promise to keep the dialog open until it
   * settles: Cancel is disabled, and the confirm button turns busy with a
   * spinner and `pendingLabel`. It closes on success and stays open if the
   * promise rejects.
   */
  onConfirm?: () => void | Promise<unknown>;
  onCancel?: () => void;
  /** Confirm-button label while an async `onConfirm` runs. */
  pendingLabel?: ReactNode;
  /** Disables the confirm action, e.g. until a confirmation phrase is typed. */
  isConfirmDisabled?: boolean;
  /** Close on a backdrop press. Off by default: an alert asks for an answer. */
  isDismissable?: boolean;
  /** Extra content between the description and the actions. */
  children?: ReactNode;
}

/** Spinner sized for Button's icon slot, in the button's own text colour. */
function PendingGlyph({ className }: { className?: string }) {
  return <SpinnerGlyph size="sm" className={cx(className, "text-current")} />;
}

export function AlertDialog({
  title,
  description,
  variant = "danger",
  icon = RiErrorWarningLine,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  onConfirm,
  onCancel,
  pendingLabel,
  isConfirmDisabled = false,
  isDismissable = false,
  size = "sm",
  children,
  ...props
}: AlertDialogProps) {
  const [pending, setPending] = useState(false);

  return (
    <Modal {...props} size={size} role="alertdialog" isDismissable={isDismissable && !pending} isKeyboardDismissDisabled={pending}>
      {({ close }) => {
        const confirm = async () => {
          if (pending) return;
          const result = onConfirm?.();
          if (!result || typeof result.then !== "function") return close();
          setPending(true);
          try {
            await result;
            close();
          } catch {
            // Stay open: the caller reports the failure.
          } finally {
            setPending(false);
          }
        };

        return (
          <>
            <ModalHeader
              title={title}
              description={description}
              icon={icon || undefined}
              tone={variant === "danger" ? "danger" : "default"}
              showCloseButton={false}
            />
            {children ? <ModalBody>{children}</ModalBody> : null}
            <ModalFooter>
              <Button
                variant="secondary"
                autoFocus
                disabled={pending}
                onClick={() => {
                  onCancel?.();
                  close();
                }}
              >
                {cancelLabel}
              </Button>
              <Button
                variant={variant === "danger" ? "danger" : "primary"}
                // Busy rather than disabled while pending: the button keeps
                // its colour, so the spinner on it stays legible.
                disabled={isConfirmDisabled}
                aria-busy={pending || undefined}
                leadingIcon={pending ? PendingGlyph : undefined}
                onClick={confirm}
              >
                {pending && pendingLabel ? pendingLabel : confirmLabel}
              </Button>
            </ModalFooter>
          </>
        );
      }}
    </Modal>
  );
}
