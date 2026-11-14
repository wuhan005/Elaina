/**
 * Shared visual recipe for BoardUI menu surfaces.
 *
 * Select and Dropdown intentionally keep their distinct React Aria semantics
 * (single-value listbox vs. action menu), while consuming the same panel,
 * motion, and row treatments from here.
 */
export const MENU_POPOVER_SURFACE = [
  // z-110 sits above the modal layer (z-100), so menus opened inside a
  // Modal or Drawer render over it rather than behind its backdrop.
  "z-110 max-w-[calc(100vw-32px)] overflow-y-auto",
  "rounded-2xl border border-border-button-default bg-background-primary-default p-2.5 shadow-dropdown",
  "transition duration-150 ease-out",
  "data-[entering]:opacity-0 data-[entering]:scale-95 data-[entering]:blur-[2px]",
  "data-[exiting]:opacity-0 data-[exiting]:scale-95 data-[exiting]:blur-[2px]",
  "data-[placement=bottom]:origin-top-left rtl:data-[placement=bottom]:origin-top-right data-[placement=top]:origin-bottom-left rtl:data-[placement=top]:origin-bottom-right",
  "data-[placement=left]:origin-right data-[placement=right]:origin-left",
].join(" ");

export const MENU_POPOVER_WIDTH = "w-[266px]";

/**
 * BoardUI's compact dropdown (the project board's and the chat's menus, and
 * the canvas editor's app menus): a 14px panel at 4px padding over
 * px-2 py-1.5 rows. Layered over `MENU_POPOVER_SURFACE` and `MENU_ITEM`; the
 * rows' 10px radius sits concentric 4px inside the panel's 14px.
 */
export const MENU_POPOVER_COMPACT = "rounded-[14px] p-1";

export const MENU_ITEM_COMPACT = "px-2 py-1.5";

/**
 * `DropdownDivider` bleeds through the default p-2.5 panel; in the compact
 * panel it bleeds through p-1 and sits in the rows' 4px gap on each side.
 */
export const MENU_DIVIDER_COMPACT = "-mx-1 my-0";

export const MENU_ITEMS_CONTAINER = "flex w-full flex-col gap-1 outline-none";

export const MENU_ITEM = [
  "flex w-full cursor-pointer items-center gap-2 rounded-2lg p-2 text-start",
  "text-text-primary outline-none transition-colors",
].join(" ");

export const MENU_ITEM_ACTIVE = "bg-dropdown-item-hover-background";

export const MENU_ITEM_INTERACTIVE =
  "hover:bg-dropdown-item-hover-background focus-visible:bg-dropdown-item-hover-background";
