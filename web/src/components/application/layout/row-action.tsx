import type { ComponentType } from "react";
import { Focusable } from "react-aria-components";
import { IconButton, IconLinkButton } from "@/components/base/buttons/icon-button";
import { Tooltip, TooltipTrigger } from "@/components/base/tooltip/tooltip";
import { cx } from "@/utils/cx";

type IconComponent = ComponentType<{ className?: string; "aria-hidden"?: boolean | "true" | "false" }>;

// RowAction 是表格行内带 Tooltip 的图标操作按钮，传 href 时渲染为链接。
export function RowAction({
  icon,
  label,
  onClick,
  href,
  danger = false,
}: {
  icon: IconComponent;
  label: string;
  onClick?: () => void;
  href?: string;
  danger?: boolean;
}) {
  const className = cx(danger && "hover:text-text-error-primary");
  return (
    <TooltipTrigger delay={200}>
      <Focusable>
        {href ? (
          <IconLinkButton icon={icon} size="small" aria-label={label} href={href} target="_blank" rel="noreferrer" className={className} />
        ) : (
          <IconButton icon={icon} size="small" aria-label={label} onClick={onClick} className={className} />
        )}
      </Focusable>
      <Tooltip size="md">{label}</Tooltip>
    </TooltipTrigger>
  );
}
