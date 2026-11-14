import type { ComponentType, ReactNode } from "react";
import { Breadcrumb, BreadcrumbItem } from "@/components/base/breadcrumb/breadcrumb";
import { useT } from "@/i18n";
import { cx } from "@/utils/cx";

type IconComponent = ComponentType<{ className?: string; "aria-hidden"?: boolean | "true" | "false" }>;

export interface Crumb {
  label: string;
  href?: string;
  icon?: IconComponent;
}

export function PageHeader({
  title,
  description,
  breadcrumbs,
  actions,
  className,
}: {
  title: ReactNode;
  description?: ReactNode;
  breadcrumbs?: Crumb[];
  actions?: ReactNode;
  className?: string;
}) {
  const t = useT();
  return (
    <header className={cx("flex flex-col gap-3 pb-6", className)}>
      {breadcrumbs && breadcrumbs.length > 0 && (
        <Breadcrumb className="-ms-1" aria-label={t("Breadcrumb")}>
          {breadcrumbs.map((crumb, index) => (
            <BreadcrumbItem
              key={crumb.label}
              href={crumb.href}
              icon={crumb.icon}
              current={index === breadcrumbs.length - 1}
            >
              {crumb.label}
            </BreadcrumbItem>
          ))}
        </Breadcrumb>
      )}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="flex min-w-0 flex-col gap-1">
          <h1 className="text-title-1-medium text-text-primary">{title}</h1>
          {description && <p className="text-body-regular text-text-secondary">{description}</p>}
        </div>
        {actions && <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div>}
      </div>
    </header>
  );
}
