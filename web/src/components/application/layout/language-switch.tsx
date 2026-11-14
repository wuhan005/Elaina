import { RiTranslate2 } from "@remixicon/react";
import { SegmentedControl, SegmentedControlItem } from "@/components/base/segmented-control/segmented-control";
import { setLocale, useLocale, useT, type Locale } from "@/i18n";
import { cx } from "@/utils/cx";

export function LanguageSwitch({ collapsed = false, className }: { collapsed?: boolean; className?: string }) {
  const locale = useLocale();
  const t = useT();

  if (collapsed) {
    const next: Locale = locale === "zh-CN" ? "en" : "zh-CN";
    return (
      <button
        type="button"
        aria-label={t("Switch language")}
        title={next === "en" ? "English" : "简体中文"}
        onClick={() => setLocale(next)}
        className={cx(
          "flex size-9 cursor-pointer items-center justify-center rounded-2lg text-foreground-icon-secondary outline-none",
          "transition-colors duration-150 ease hover:bg-background-secondary-hover hover:text-foreground-icon-primary",
          "focus-visible:ring-2 focus-visible:ring-border-focus-ring",
          className,
        )}
      >
        <RiTranslate2 className="size-5" aria-hidden />
      </button>
    );
  }

  return (
    <SegmentedControl
      aria-label={t("Language")}
      selectedKeys={new Set([locale])}
      onSelectionChange={(keys) => {
        const next = [...keys][0];
        if (next) setLocale(next as Locale);
      }}
      className={className}
    >
      <SegmentedControlItem id="en" aria-label="English">
        EN
      </SegmentedControlItem>
      <SegmentedControlItem id="zh-CN" aria-label="简体中文">
        中文
      </SegmentedControlItem>
    </SegmentedControl>
  );
}
