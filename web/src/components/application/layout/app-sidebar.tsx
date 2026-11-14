import { useEffect, useState, type ComponentType, type ReactNode } from "react";
import { Link, NavLink } from "react-router";
import {
  RiCodeBoxLine,
  RiDashboardHorizontalLine,
  RiExternalLinkLine,
  RiGithubLine,
  RiLogoutBoxRLine,
  RiSideBarFill,
  RiStackLine,
} from "@remixicon/react";
import { client } from "@/api/client";
import { ThemeToggle } from "@/components/application/theme/theme-toggle";
import { Avatar } from "@/components/base/avatar/avatar";
import {
  Dropdown,
  DropdownDivider,
  DropdownItem,
  DropdownPopover,
  DropdownTrigger,
} from "@/components/base/dropdown/dropdown";
import { toast } from "@/components/base/toast/toast";
import { ChevronUpDownSmall } from "@/components/foundations/icons/chevrons";
import { authStore } from "@/store/auth";
import { cx } from "@/utils/cx";
import { ElainaMark } from "./elaina-mark";

type IconComponent = ComponentType<{ className?: string; "aria-hidden"?: boolean | "true" | "false" }>;

interface NavEntry {
  to: string;
  label: string;
  icon: IconComponent;
}

const NAV: NavEntry[] = [
  { to: "/dashboard", label: "Dashboard", icon: RiDashboardHorizontalLine },
  { to: "/template", label: "Templates", icon: RiStackLine },
  { to: "/sandbox", label: "Sandboxes", icon: RiCodeBoxLine },
];

const GITHUB_URL = "https://github.com/wuhan005/Elaina";
const COLLAPSED_STORAGE_KEY = "elaina:sidebar-collapsed";

// 侧边栏收起时标签以模糊 + 淡出的方式隐藏，图标位置保持不动。
function Collapsible({ collapsed, children, className }: { collapsed: boolean; children: ReactNode; className?: string }) {
  return (
    <span
      className={cx(
        "flex min-w-0 items-center overflow-hidden transition-[max-width,opacity,filter] duration-300 ease-in-out",
        collapsed ? "max-w-0 opacity-0 blur-[3px]" : "max-w-full opacity-100 blur-0",
        className,
      )}
    >
      {children}
    </span>
  );
}

const navItemClass = (collapsed: boolean, selected: boolean) =>
  cx(
    "flex items-center overflow-hidden rounded-2lg p-2 outline-none",
    "transition-[width,background-color] duration-300 ease-in-out",
    "focus-visible:ring-2 focus-visible:ring-border-focus-ring",
    collapsed ? "w-9" : "w-full",
    selected
      ? "bg-linear-to-b from-accent-500 to-accent-600 shadow-nav-selected"
      : "hover:bg-background-secondary-hover",
  );

function NavItem({ to, label, icon: Icon, collapsed }: NavEntry & { collapsed: boolean }) {
  return (
    <NavLink
      to={to}
      aria-label={label}
      title={collapsed ? label : undefined}
      className={({ isActive }) => navItemClass(collapsed, isActive)}
    >
      {({ isActive }) => (
        <span className="flex min-w-0 items-center gap-2">
          <Icon
            className={cx("size-5 shrink-0", isActive ? "text-text-white" : "text-foreground-icon-secondary")}
            aria-hidden
          />
          <Collapsible collapsed={collapsed}>
            <span className={cx("text-body-medium whitespace-nowrap", isActive ? "text-text-white" : "text-text-secondary")}>
              {label}
            </span>
          </Collapsible>
        </span>
      )}
    </NavLink>
  );
}

function ExternalNavItem({ href, label, icon: Icon, collapsed }: { href: string; label: string; icon: IconComponent; collapsed: boolean }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      aria-label={label}
      title={collapsed ? label : undefined}
      className={cx(navItemClass(collapsed, false), "group justify-between")}
    >
      <span className="flex min-w-0 items-center gap-2">
        <Icon className="size-5 shrink-0 text-foreground-icon-secondary" aria-hidden />
        <Collapsible collapsed={collapsed}>
          <span className="text-body-medium whitespace-nowrap text-text-secondary">{label}</span>
        </Collapsible>
      </span>
      <Collapsible collapsed={collapsed}>
        <RiExternalLinkLine
          className="size-4 shrink-0 text-foreground-icon-quaternary opacity-0 transition-opacity group-hover:opacity-100"
          aria-hidden
        />
      </Collapsible>
    </a>
  );
}

function AccountMenu({ collapsed }: { collapsed: boolean }) {
  const [isOpen, setIsOpen] = useState(false);

  const signOut = () => {
    setIsOpen(false);
    client.auth
      .signOut()
      .then((res) => toast.success(res.data.data))
      .catch(() => undefined)
      .finally(() => authStore.clear());
  };

  return (
    <Dropdown isOpen={isOpen} onOpenChange={setIsOpen}>
      <DropdownTrigger
        aria-label="Account menu"
        className={cx(
          "flex items-center justify-between gap-2 rounded-2lg text-start",
          "transition-[width,background-color] duration-300 ease-in-out hover:bg-background-secondary-hover",
          collapsed ? "w-9 justify-center p-0.5" : "w-full p-1.5",
          isOpen && "bg-background-secondary-hover",
        )}
      >
        <span className="flex min-w-0 items-center gap-2">
          <Avatar size="md" color="blue" initials="A" />
          <Collapsible collapsed={collapsed}>
            <span className="flex min-w-0 flex-col">
              <span className="truncate text-body-medium text-text-primary">Administrator</span>
              <span className="truncate text-caption-1-medium text-text-tertiary">Signed in</span>
            </span>
          </Collapsible>
        </span>
        <Collapsible collapsed={collapsed}>
          <ChevronUpDownSmall className="size-4 shrink-0 text-foreground-icon-tertiary" />
        </Collapsible>
      </DropdownTrigger>
      <DropdownPopover aria-label="Account" placement={collapsed ? "right bottom" : "top start"} offset={8} className="w-[232px]">
        <div className="flex items-center gap-2 px-2 pt-1 pb-0.5">
          <Avatar size="sm" color="blue" initials="A" />
          <span className="flex min-w-0 flex-col">
            <span className="truncate text-body-medium text-text-primary">Administrator</span>
            <span className="truncate text-caption-1-medium text-text-tertiary">Elaina console</span>
          </span>
        </div>
        <DropdownDivider />
        <DropdownItem onSelect={signOut}>
          <RiLogoutBoxRLine className="size-[18px] shrink-0 text-foreground-icon-secondary" aria-hidden />
          <span className="text-body-medium text-text-primary">Sign out</span>
        </DropdownItem>
      </DropdownPopover>
    </Dropdown>
  );
}

// 小屏下强制收起为图标栏。
function useIsCompactViewport() {
  const query = "(max-width: 1023px)";
  const [compact, setCompact] = useState(() => window.matchMedia(query).matches);
  useEffect(() => {
    const mq = window.matchMedia(query);
    const onChange = (event: MediaQueryListEvent) => setCompact(event.matches);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);
  return compact;
}

export function AppSidebar({ className }: { className?: string }) {
  const compactViewport = useIsCompactViewport();
  const [collapsedState, setCollapsedState] = useState(() => window.localStorage.getItem(COLLAPSED_STORAGE_KEY) === "1");
  const collapsed = compactViewport || collapsedState;

  const toggle = () => {
    const next = !collapsedState;
    setCollapsedState(next);
    window.localStorage.setItem(COLLAPSED_STORAGE_KEY, next ? "1" : "0");
  };

  return (
    <aside
      className={cx(
        "flex h-full shrink-0 flex-col justify-between overflow-hidden",
        "rounded-3xl border border-border-button-white bg-background-secondary-default shadow-sidebar",
        "transition-[width] duration-300 ease-in-out",
        collapsed ? "w-[60px] px-[11px] py-3" : "w-[248px] p-3",
        className,
      )}
    >
      <div className="-m-2 flex min-h-0 w-[calc(100%+16px)] flex-col gap-6 overflow-y-auto p-2 [scrollbar-width:none]">
        {/* 品牌区与收起按钮 */}
        <div
          className={cx(
            "flex w-full transition-[gap] duration-300 ease-in-out",
            collapsed ? "flex-col items-start gap-3" : "flex-row items-center justify-between",
          )}
        >
          <Link to="/dashboard" aria-label="Elaina home" className="flex min-w-0 items-center gap-2.5 rounded-2lg outline-none focus-visible:ring-2 focus-visible:ring-border-focus-ring">
            <ElainaMark />
            <Collapsible collapsed={collapsed}>
              <span className="flex min-w-0 flex-col">
                <span className="text-headline-semibold whitespace-nowrap text-text-primary">Elaina</span>
                <span className="text-caption-1-medium whitespace-nowrap text-text-tertiary">WASM code runner</span>
              </span>
            </Collapsible>
          </Link>
          {!compactViewport && (
            <button
              type="button"
              aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
              aria-expanded={!collapsed}
              onClick={toggle}
              className={cx(
                "cursor-pointer rounded-md text-foreground-icon-secondary outline-none transition-colors hover:text-foreground-icon-primary",
                "focus-visible:ring-2 focus-visible:ring-border-focus-ring",
                collapsed && "flex w-9 items-center justify-center",
              )}
            >
              <RiSideBarFill
                className={cx("size-5 transition-transform duration-300 ease-in-out", collapsed ? "rtl:-scale-x-100" : "-scale-x-100 rtl:scale-x-100")}
                aria-hidden
              />
            </button>
          )}
        </div>

        <nav className={cx("flex w-full flex-col gap-1", !collapsed && "px-0.5")}>
          <Collapsible collapsed={collapsed} className="mb-1 ps-2">
            <span className="text-caption-1-medium whitespace-nowrap text-text-tertiary">Workspace</span>
          </Collapsible>
          {NAV.map((item) => (
            <NavItem key={item.to} {...item} collapsed={collapsed} />
          ))}
        </nav>
      </div>

      <div className="flex w-full shrink-0 flex-col gap-3">
        {collapsed ? <ThemeToggle collapsed /> : <ThemeToggle appearance="sidebar-segmented" />}
        <nav className="flex w-full flex-col gap-1">
          <ExternalNavItem href={GITHUB_URL} label="GitHub" icon={RiGithubLine} collapsed={collapsed} />
        </nav>
        <div className="h-px w-full bg-separator-border" />
        <AccountMenu collapsed={collapsed} />
      </div>
    </aside>
  );
}
