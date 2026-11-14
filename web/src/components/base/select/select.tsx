"use client";

import { createContext, useContext, useLayoutEffect, useRef, useState } from "react";
import type { ReactNode, Ref } from "react";
import {
  Button as AriaButton,
  ListBox as AriaListBox,
  ListBoxItem as AriaListBoxItem,
  Popover as AriaPopover,
  Select as AriaSelect,
  SelectValue as AriaSelectValue,
} from "react-aria-components";
import type {
  ListBoxItemProps as AriaListBoxItemProps,
  SelectProps as AriaSelectProps,
  SelectValueRenderProps as AriaSelectValueRenderProps,
} from "react-aria-components";
import {
  MENU_ITEM,
  MENU_ITEM_ACTIVE,
  MENU_ITEM_COMPACT,
  MENU_ITEMS_CONTAINER,
  MENU_POPOVER_COMPACT,
  MENU_POPOVER_SURFACE,
  MENU_POPOVER_WIDTH,
} from "@/components/base/dropdown/menu-styles";
import { Label } from "@/components/base/input/label";
import { ScrollFade, useScrollFade } from "@/components/foundations/scroll-fade/scroll-fade";
import { ChevronDownSmall } from "@/components/foundations/icons/chevrons";
import { cx } from "@/utils/cx";
import { useDismissOnOutsidePress, useTriggerToggle } from "@/utils/use-dismiss-on-outside-press";

/**
 * Figma source: Board UI → dashboard 1 dropdown triggers ("All prices" filter
 * node 3731:3208, table status dropdown node 3731:3264).
 *
 * Select with a button trigger styled 1:1 with Figma:
 *   trigger  bg color/white, 1px border/button/default, shadow/xs,
 *            px 10 py 8, radius/2lg (10px), gap 6,
 *            Body 1/Medium text color/neutral/950,
 *            16px chevron (custom stroke glyph) in text/secondary
 *
 * The open popover/listbox shares the BoardUI menu recipe with Dropdown
 * (surface, shadow/dropdown, 150ms fade/scale/blur motion, row hover and
 * selected surface) in one of two densities:
 *   compact      (default) BoardUI's compact dropdown, as the canvas editor's
 *                menus wear it: a 14px panel at 4px padding over px-2 py-1.5
 *                rows 4px apart, as wide as the trigger (or its longest row)
 *   comfortable  the roomy menu: a 266px radius/2xl panel at 8px padding
 *                over p-2 rows
 *
 * A long list scrolls inside the panel under the progressive top blur
 * (ScrollFade), and rows scrolled in from above land below it.
 *
 * Item content is free-form: pass a `StatusDot` + text for status selects.
 *
 * `label` renders Input's field label above the trigger through React Aria's
 * Label, which names the select for assistive technology. Without it, pass
 * `aria-label` (or `aria-labelledby`) instead. `placeholder` is React Aria's:
 * the trigger shows it until something is selected.
 */

export type SelectSize = "sm" | "md";

export type SelectVariant = "compact" | "comfortable";

const SelectStyleContext = createContext<{ size: SelectSize; variant: SelectVariant }>({
  size: "md",
  variant: "compact",
});

export interface SelectProps<T extends object>
  extends Omit<AriaSelectProps<T>, "children"> {
  /** Visible label above the trigger; React Aria wires it as the select's accessible name. */
  label?: ReactNode;
  /** Trigger width. Defaults to hug content; the compact menu is at least as wide as the trigger. */
  className?: string;
  triggerClassName?: string;
  /** Classes for the open popover (e.g. constrain its width). */
  popoverClassName?: string;
  /** `md` (default) or `sm` for compact/dense contexts (e.g. compact tables). */
  size?: SelectSize;
  /** The open menu's density: `compact` (default) or the roomier `comfortable`. */
  variant?: SelectVariant;
  children: ReactNode;
  items?: Iterable<T>;
  /** Customise the trigger's rendered value (e.g. a compact flag + dial code).
   *  Falls back to the selected item's own content when omitted. */
  renderValue?: ReactNode | ((values: AriaSelectValueRenderProps<T>) => ReactNode);
  ref?: Ref<HTMLDivElement>;
}

export function Select<T extends object>({
  label,
  className,
  triggerClassName,
  popoverClassName,
  size = "md",
  variant = "compact",
  children,
  items,
  renderValue,
  ref,
  ...props
}: SelectProps<T>) {
  // The popover is `isNonModal`: react-aria's modal scroll lock puts
  // `overflow: hidden` on <html>, which — with this app's `h-full` root —
  // collapses the page scroll position and visibly yanks the sticky
  // sidebar upward every time a select opens. Non-modal skips the lock,
  // but also disables react-aria's outside-press dismissal (`isDismissable`
  // is hard-coupled to `!isNonModal` in usePopover), so open state is
  // controlled here and dismissal restored manually — the same fix as the
  // date-picker family (see utils/use-dismiss-on-outside-press.ts).
  const triggerRef = useRef<HTMLButtonElement>(null);
  const popoverRef = useRef<HTMLElement>(null);
  const [isOpen, setIsOpen] = useState(false);
  useDismissOnOutsidePress(isOpen, () => setIsOpen(false), [triggerRef, popoverRef]);
  // Pressing the trigger while open closes the popover instead of reopening
  const allowOpenChange = useTriggerToggle(isOpen, triggerRef);
  const compact = variant === "compact";

  return (
    <AriaSelect
      ref={ref}
      {...props}
      isOpen={isOpen}
      onOpenChange={(o) => allowOpenChange(o) && setIsOpen(o)}
      className={cx("group flex flex-col", label ? "gap-1" : undefined, className)}
    >
      {({ isOpen, isRequired }) => (
        <>
          {label ? <Label isRequired={isRequired}>{label}</Label> : null}
          <AriaButton
            ref={triggerRef}
            className={cx(
              "flex w-full cursor-pointer items-center justify-between rounded-2lg",
              "border border-border-button-default bg-background-primary-default shadow-xs",
              "text-text-primary",
              "transition-[background-color,border-color,box-shadow,padding,font-size] duration-200 ease",
              "hover:bg-background-primary-hover hover:border-border-button-hover",
              "outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-border-focus-ring",
              "disabled:cursor-not-allowed disabled:bg-background-primary-disabled disabled:text-text-tertiary disabled:shadow-none",
              size === "sm"
                ? "gap-1 px-[7px] py-1 text-body-2-medium"
                : "gap-1.5 px-2.5 py-2 text-body-medium",
              triggerClassName,
            )}
          >
            <AriaSelectValue
              className={cx(
                "flex min-w-0 items-center truncate",
                size === "sm" ? "gap-1" : "gap-[5px]",
              )}
            >
              {renderValue as never /* RAC accepts node or render-prop */}
            </AriaSelectValue>
            <ChevronDownSmall
              className={cx(
                "shrink-0 text-text-secondary transition-transform duration-200 ease",
                size === "sm" ? "size-3.5" : "size-4",
                isOpen && "rotate-180",
              )}
            />
          </AriaButton>
          <AriaPopover
            ref={popoverRef}
            isNonModal
            offset={4}
            className={cx(
              MENU_POPOVER_SURFACE,
              "flex flex-col",
              compact
                ? // As wide as the trigger, or as its longest row.
                  [MENU_POPOVER_COMPACT, "w-max min-w-(--trigger-width)"]
                : // Listbox rows already space themselves 4px apart, so the
                  // surface sits a touch tighter than the action-menu Dropdown.
                  [MENU_POPOVER_WIDTH, "p-2"],
              popoverClassName,
            )}
          >
            <SelectStyleContext.Provider value={{ size, variant }}>
              <SelectList items={items}>{children}</SelectList>
            </SelectStyleContext.Provider>
          </AriaPopover>
        </>
      )}
    </AriaSelect>
  );
}

/**
 * The listbox and its top blur. Mounted with the popover, so every opening
 * starts with a fresh fade, and a list opened at its top shows none.
 *
 * A long list opens with its selected row centred (a year list opens on the
 * chosen year, not on its first). React Aria's own scroll on open does not
 * run for a pointer open here; it still keeps the keyboard's row in view.
 */
function SelectList<T extends object>({ items, children }: { items?: Iterable<T>; children: ReactNode }) {
  const { size, variant } = useContext(SelectStyleContext);
  const fade = useScrollFade();
  const listRef = useRef<HTMLDivElement>(null);
  useLayoutEffect(() => {
    const list = listRef.current;
    const selected = list?.querySelector<HTMLElement>('[role="option"][aria-selected="true"]');
    if (!list || !selected || list.scrollHeight <= list.clientHeight) return;
    // Layout offsets, not client rects: the panel is still scaled while it enters.
    const top = selected.offsetTop - list.offsetTop;
    list.scrollTop = top - (list.clientHeight - selected.offsetHeight) / 2;
  }, []);
  return (
    <div className="relative flex min-h-0 flex-col">
      <AriaListBox
        ref={listRef}
        items={items}
        onScroll={fade.onScroll}
        className={cx(
          MENU_ITEMS_CONTAINER,
          "min-h-0 overflow-y-auto overscroll-contain scroll-pt-6",
          // Six rows and half of the seventh, so a longer list reads as one
          // that scrolls; the comfortable menu keeps its 240px.
          variant === "comfortable" ? "max-h-[240px]" : size === "sm" ? "max-h-[219px]" : "max-h-[232px]",
        )}
      >
        {children}
      </AriaListBox>
      <ScrollFade progress={fade.progress} />
    </div>
  );
}

export interface SelectItemProps extends Omit<AriaListBoxItemProps, "children"> {
  children?: ReactNode;
}

export function SelectItem({ className, children, ...props }: SelectItemProps) {
  const { size, variant } = useContext(SelectStyleContext);
  return (
    <AriaListBoxItem
      {...props}
      className={(state) =>
        cx(
          MENU_ITEM,
          size === "sm"
            ? [MENU_ITEM_COMPACT, "text-body-2-medium"]
            : [variant === "compact" && MENU_ITEM_COMPACT, "text-body-medium"],
          (state.isFocused || state.isSelected) && MENU_ITEM_ACTIVE,
          state.isDisabled && "cursor-not-allowed text-text-disabled",
          typeof className === "function" ? className(state) : className,
        )
      }
    >
      {children}
    </AriaListBoxItem>
  );
}
