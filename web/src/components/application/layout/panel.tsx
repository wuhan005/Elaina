import type { ReactNode } from "react";
import { cx } from "@/utils/cx";

// Panel is a card container with an optional header.
export function Panel({
  title,
  description,
  actions,
  children,
  className,
  bodyClassName,
}: {
  title?: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
  bodyClassName?: string;
}) {
  const hasHeader = title !== undefined || actions !== undefined;
  return (
    <section
      className={cx(
        "flex min-w-0 flex-col rounded-3xl border border-border-button-default bg-background-full shadow-xs",
        className,
      )}
    >
      {hasHeader && (
        <div className="flex items-start justify-between gap-3 px-5 pt-5">
          <div className="flex min-w-0 flex-col gap-0.5">
            {title && <h2 className="text-headline-medium text-text-primary">{title}</h2>}
            {description && <p className="text-body-regular text-text-secondary">{description}</p>}
          </div>
          {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
        </div>
      )}
      <div className={cx("flex min-w-0 flex-col p-5", bodyClassName)}>{children}</div>
    </section>
  );
}

// FormSection is a form row with the title on the left and fields on the right.
export function FormSection({
  title,
  description,
  children,
}: {
  title: ReactNode;
  description?: ReactNode;
  children: ReactNode;
}) {
  return (
    <div className="grid gap-4 border-b border-separator-border px-5 py-6 last:border-b-0 md:grid-cols-[minmax(0,260px)_minmax(0,1fr)] md:gap-10 md:px-6">
      <div className="flex flex-col gap-1">
        <h2 className="text-headline-medium text-text-primary">{title}</h2>
        {description && <p className="text-body-regular text-text-secondary">{description}</p>}
      </div>
      <div className="flex min-w-0 flex-col gap-5">{children}</div>
    </div>
  );
}

// EmptyState is the placeholder for an empty list.
export function EmptyState({
  icon,
  title,
  description,
  action,
}: {
  icon: ReactNode;
  title: ReactNode;
  description?: ReactNode;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 px-6 py-14 text-center">
      <span className="flex size-12 items-center justify-center rounded-2xl bg-background-tertiary-default text-foreground-icon-secondary">
        {icon}
      </span>
      <div className="flex max-w-sm flex-col gap-1">
        <p className="text-headline-medium text-text-primary">{title}</p>
        {description && <p className="text-body-regular text-text-secondary">{description}</p>}
      </div>
      {action && <div className="mt-2">{action}</div>}
    </div>
  );
}
