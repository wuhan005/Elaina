"use client";

import type { ComponentType, ReactNode } from "react";
import { Button as AriaButton, NumberField as AriaNumberField } from "react-aria-components";
import type { NumberFieldProps as AriaNumberFieldProps } from "react-aria-components";
import { RiAddLine, RiSubtractLine } from "@remixicon/react";
import { HintText } from "@/components/base/input/hint-text";
import { InputBase } from "@/components/base/input/input";
import type { InputSize } from "@/components/base/input/input";
import { Label } from "@/components/base/input/label";
import { cx } from "@/utils/cx";

/**
 * Number field — a number typed or stepped: seats, a price, a percentage,
 * a weight.
 *
 *   <NumberField label="Seats" defaultValue={5} minValue={1} maxValue={50} />
 *   <NumberField label="Price" defaultValue={49} formatOptions={{ style: "currency", currency: "USD" }} />
 *
 * Input's field (surface, hover and focus rings, invalid and disabled
 * states, two sizes) with minus and plus steppers at its end. Built on React
 * Aria's NumberField, so the value is formatted and parsed for the locale
 * (currency, percent and unit styles through `formatOptions`), stays within
 * `minValue` and `maxValue` on blur, and the steppers switch off at either
 * limit. Up and Down step it, Home and End jump to the limits, the scroll
 * wheel steps it while it has focus (`isWheelDisabled` turns that off), and
 * holding a stepper repeats it.
 */

type IconComponent = ComponentType<{ className?: string; "aria-hidden"?: boolean | "true" | "false" }>;

export interface NumberFieldProps extends Omit<AriaNumberFieldProps, "className" | "children"> {
  label?: ReactNode;
  hint?: ReactNode;
  /** Show an info icon next to the label. */
  tooltip?: boolean | string;
  placeholder?: string;
  /** 36px (default) or 32px. */
  size?: InputSize;
  leadingIcon?: IconComponent;
  /** Hide the minus and plus buttons; the keys still step the value. */
  hideSteppers?: boolean;
  className?: string;
  fieldClassName?: string;
}

const STEPPER = [
  "flex shrink-0 cursor-pointer items-center justify-center rounded-md text-foreground-icon-secondary outline-none",
  "transition-colors duration-150 ease-out hover:bg-background-tertiary-hover hover:text-text-primary",
  "data-[pressed]:bg-background-tertiary-hover data-[focus-visible]:ring-2 data-[focus-visible]:ring-border-focus-ring",
  "data-[disabled]:cursor-not-allowed data-[disabled]:text-input-disabled-text data-[disabled]:hover:bg-transparent",
].join(" ");

export function NumberField({
  label,
  hint,
  tooltip,
  placeholder,
  size = "medium",
  leadingIcon,
  hideSteppers = false,
  "aria-label": ariaLabel,
  className,
  fieldClassName,
  ...props
}: NumberFieldProps) {
  const stepper = cx(STEPPER, size === "small" ? "size-5" : "size-6");

  return (
    <AriaNumberField
      {...props}
      aria-label={ariaLabel ?? (!label ? placeholder : undefined)}
      className={cx("group flex w-full flex-col items-start gap-1", className)}
    >
      {({ isInvalid }) => (
        <>
          {label ? (
            <Label isRequired={props.isRequired} isInvalid={isInvalid} tooltip={tooltip}>
              {label}
            </Label>
          ) : null}
          <InputBase
            size={size}
            placeholder={placeholder}
            leadingIcon={leadingIcon}
            fieldClassName={fieldClassName}
            className="tabular-nums"
            trailingAddon={
              hideSteppers ? undefined : (
                <span className="-my-1 -me-1 flex shrink-0 items-center gap-0.5">
                  <AriaButton slot="decrement" className={stepper}>
                    <RiSubtractLine aria-hidden className="size-4" />
                  </AriaButton>
                  <AriaButton slot="increment" className={stepper}>
                    <RiAddLine aria-hidden className="size-4" />
                  </AriaButton>
                </span>
              )
            }
          />
          {hint ? <HintText isInvalid={isInvalid}>{hint}</HintText> : null}
        </>
      )}
    </AriaNumberField>
  );
}
