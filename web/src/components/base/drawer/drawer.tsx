"use client";

import type { ReactNode } from "react";
import { Modal as AriaModal, ModalOverlay as AriaModalOverlay } from "react-aria-components";
import type { DialogRenderProps, ModalOverlayProps } from "react-aria-components";
import {
  ModalBody,
  ModalDialog,
  ModalFooter,
  ModalHeader,
  ModalTrigger,
} from "@/components/base/modal/modal";
import { useDirection } from "@/components/foundations/direction/direction";
import { cx, sortCx } from "@/utils/cx";

/**
 * Drawer — a panel that slides in from an edge of the screen over a dimmed
 * page. It is Modal's sibling on the same React Aria ModalOverlay, so focus
 * is trapped and returns to the trigger, the page stops scrolling, and
 * Escape or a press on the backdrop closes it. The header, body, footer,
 * and trigger are Modal's, re-exported under Drawer names.
 *
 *   <DrawerTrigger>
 *     <Button>Filters</Button>
 *     <Drawer placement="end">
 *       {({ close }) => (
 *         <>
 *           <DrawerHeader title="Filters" />
 *           <DrawerBody>…</DrawerBody>
 *           <DrawerFooter>…</DrawerFooter>
 *         </>
 *       )}
 *     </Drawer>
 *   </DrawerTrigger>
 *
 * The panel floats 12px inside the viewport (8px on phones) at radius/3xl,
 * the same surface as Modal and the floating sidebar. `start` and `end`
 * follow the reading direction, so an end drawer opens from the left in RTL.
 *
 * Motion: the panel slides in from past its edge and sharpens from a 4px
 * blur over 300ms on the iOS sheet curve (Modal's timing), and slides back
 * out the same way; the black/40 backdrop fades with it. Reduced motion
 * swaps the slide for a fade.
 */

export type DrawerPlacement = "start" | "end" | "top" | "bottom";
export type DrawerSize = "sm" | "md" | "lg";

const styles = sortCx({
  overlay: "group/drawer fixed inset-x-0 top-0 z-100 flex h-(--visual-viewport-height) p-2 sm:p-3",
  backdrop: [
    "absolute inset-0 bg-black/40",
    "transition-opacity duration-300 ease-out",
    "group-data-[entering]/drawer:opacity-0 group-data-[exiting]/drawer:opacity-0",
  ].join(" "),
  modal: [
    "relative flex w-full flex-col outline-none",
    "transform-gpu transition-[translate,filter,opacity] duration-300 ease-[cubic-bezier(0.32,0.72,0,1)]",
    "motion-safe:data-[entering]:blur-[4px] motion-safe:data-[exiting]:blur-[4px]",
    "motion-reduce:data-[entering]:opacity-0 motion-reduce:data-[exiting]:opacity-0",
  ].join(" "),
  dialog: "flex-1",
  // Each panel starts 24px past its edge (the 12px inset plus its shadow),
  // so no sliver of it shows before it moves. Literal per edge and direction
  // so Tailwind can see every class.
  placement: {
    end: {
      overlay: "justify-end",
      modal: [
        "h-full",
        "motion-safe:ltr:data-[entering]:translate-x-[calc(100%+24px)] motion-safe:ltr:data-[exiting]:translate-x-[calc(100%+24px)]",
        "motion-safe:rtl:data-[entering]:-translate-x-[calc(100%+24px)] motion-safe:rtl:data-[exiting]:-translate-x-[calc(100%+24px)]",
      ].join(" "),
    },
    start: {
      overlay: "justify-start",
      modal: [
        "h-full",
        "motion-safe:ltr:data-[entering]:-translate-x-[calc(100%+24px)] motion-safe:ltr:data-[exiting]:-translate-x-[calc(100%+24px)]",
        "motion-safe:rtl:data-[entering]:translate-x-[calc(100%+24px)] motion-safe:rtl:data-[exiting]:translate-x-[calc(100%+24px)]",
      ].join(" "),
    },
    bottom: {
      overlay: "items-end",
      modal: "max-h-[85%] motion-safe:data-[entering]:translate-y-[calc(100%+24px)] motion-safe:data-[exiting]:translate-y-[calc(100%+24px)]",
    },
    top: {
      overlay: "items-start",
      modal: "max-h-[85%] motion-safe:data-[entering]:-translate-y-[calc(100%+24px)] motion-safe:data-[exiting]:-translate-y-[calc(100%+24px)]",
    },
  },
  // Side drawers only; top and bottom drawers span the screen.
  size: {
    sm: "max-w-[320px]",
    md: "max-w-[400px]",
    lg: "max-w-[600px]",
  },
});

export interface DrawerProps extends Omit<ModalOverlayProps, "children" | "className" | "style"> {
  /** The edge the panel slides in from. `start` and `end` follow the reading direction. */
  placement?: DrawerPlacement;
  /** Width of a start or end drawer: 320, 400 (default), or 600px. */
  size?: DrawerSize;
  /** Accessible name when the content has no DrawerHeader title. */
  "aria-label"?: string;
  /** Classes on the panel surface. */
  className?: string;
  /** Classes on the full-screen overlay. */
  overlayClassName?: string;
  children: ReactNode | ((renderProps: DialogRenderProps) => ReactNode);
}

export function Drawer({
  placement = "end",
  size = "md",
  "aria-label": ariaLabel,
  className,
  overlayClassName,
  isDismissable = true,
  children,
  ...props
}: DrawerProps) {
  const direction = useDirection();
  const edge = styles.placement[placement];
  const side = placement === "start" || placement === "end";

  return (
    <AriaModalOverlay
      {...props}
      dir={direction}
      isDismissable={isDismissable}
      className={cx(styles.overlay, edge.overlay, overlayClassName)}
    >
      <div aria-hidden className={styles.backdrop} />
      <AriaModal className={cx(styles.modal, edge.modal, side && styles.size[size])}>
        <ModalDialog aria-label={ariaLabel} className={cx(styles.dialog, className)}>
          {children}
        </ModalDialog>
      </AriaModal>
    </AriaModalOverlay>
  );
}

export {
  ModalBody as DrawerBody,
  ModalFooter as DrawerFooter,
  ModalHeader as DrawerHeader,
  ModalTrigger as DrawerTrigger,
};
export type {
  ModalBodyProps as DrawerBodyProps,
  ModalHeaderProps as DrawerHeaderProps,
  ModalTriggerProps as DrawerTriggerProps,
} from "@/components/base/modal/modal";
