import { useState } from "react";
import { RiEyeLine, RiEyeOffLine } from "@remixicon/react";
import { HintText } from "@/components/base/input/hint-text";
import { InputBase, TextField } from "@/components/base/input/input";
import { Label } from "@/components/base/input/label";
import { translate, useT } from "@/i18n";

// The server limits passwords by UTF-8 bytes; bcrypt handles at most 72.
export function validateNewPassword(password: string): string | undefined {
  const bytes = new TextEncoder().encode(password).length;
  if (bytes < 8 || bytes > 72) return translate("Use 8-72 characters (UTF-8 bytes).");
  return undefined;
}

export function PasswordField({
  label,
  value,
  onChange,
  error,
  hint,
  autoComplete,
  autoFocus,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  error?: string | null;
  hint?: string;
  autoComplete: "current-password" | "new-password";
  autoFocus?: boolean;
}) {
  const [visible, setVisible] = useState(false);
  const t = useT();
  return (
    <TextField
      type={visible ? "text" : "password"}
      value={value}
      onChange={onChange}
      isInvalid={!!error}
      isRequired
      autoFocus={autoFocus}
      autoComplete={autoComplete}
    >
      <Label isRequired>{label}</Label>
      <InputBase
        trailingAddon={
          <button
            type="button"
            aria-label={visible ? t("Hide password") : t("Show password")}
            onClick={() => setVisible((current) => !current)}
            className="flex shrink-0 cursor-pointer items-center rounded-md text-foreground-icon-tertiary outline-none transition-colors hover:text-foreground-icon-primary focus-visible:ring-2 focus-visible:ring-border-focus-ring"
          >
            {visible ? <RiEyeOffLine className="size-5" aria-hidden /> : <RiEyeLine className="size-5" aria-hidden />}
          </button>
        }
      />
      {error ? <HintText isInvalid>{error}</HintText> : hint ? <HintText>{hint}</HintText> : null}
    </TextField>
  );
}
