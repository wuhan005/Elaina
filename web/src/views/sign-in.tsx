import { useEffect, useRef, useState, type FormEvent } from "react";
import { Navigate, useLocation, useNavigate } from "react-router";
import axios from "axios";
import { RiArrowLeftLine, RiGithubFill } from "@remixicon/react";
import { client } from "@/api/client";
import { PasswordField } from "@/components/application/auth/password-field";
import { Logo } from "@/components/application/layout/logo";
import { ThemeToggle } from "@/components/application/theme/theme-toggle";
import { Button } from "@/components/base/buttons/button";
import { LinkButton } from "@/components/base/buttons/link-button";
import { HintText } from "@/components/base/input/hint-text";
import { Input } from "@/components/base/input/input";
import { InputOtp } from "@/components/base/input-otp/input-otp";
import { Label } from "@/components/base/input/label";
import { SpinnerGlyph } from "@/components/base/spinner/spinner";
import { LanguageSwitch } from "@/components/application/layout/language-switch";
import { useT } from "@/i18n";
import { authStore, useIsAuthenticated } from "@/store/auth";

const PendingGlyph = ({ className }: { className?: string }) => <SpinnerGlyph size="sm" className={className} />;

function PasswordStep({ onPending, onSuccess }: { onPending: (token: string) => void; onSuccess: (token: string) => void }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<{ username?: string; password?: string }>({});
  const [pending, setPending] = useState(false);
  const t = useT();

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const next = {
      username: username.trim() === "" ? t("Enter your username.") : undefined,
      password: password === "" ? t("Enter your password.") : undefined,
    };
    setErrors(next);
    if (next.username || next.password) return;

    setPending(true);
    client.auth
      .signIn({ username: username.trim(), password })
      .then((res) => {
        const { token, twoFactorRequired } = res.data.data;
        if (twoFactorRequired) onPending(token);
        else onSuccess(token);
      })
      .catch(() => setPassword(""))
      .finally(() => setPending(false));
  };

  return (
    <form className="flex flex-col gap-5" onSubmit={onSubmit} noValidate>
      <Input
        label={t("Username")}
        value={username}
        onChange={(value) => {
          setUsername(value);
          setErrors((current) => ({ ...current, username: undefined }));
        }}
        isRequired
        isInvalid={!!errors.username}
        errorMessage={errors.username}
        autoComplete="username"
        autoFocus
        inputDir="ltr"
      />
      <PasswordField
        label={t("Password")}
        value={password}
        onChange={(value) => {
          setPassword(value);
          setErrors((current) => ({ ...current, password: undefined }));
        }}
        error={errors.password}
        autoComplete="current-password"
      />
      <Button type="submit" className="mt-1 w-full" disabled={pending} aria-busy={pending || undefined} leadingIcon={pending ? PendingGlyph : undefined}>
        {t("Sign in")}
      </Button>
    </form>
  );
}

function TwoFactorStep({ token, onBack, onSuccess }: { token: string; onBack: () => void; onSuccess: (token: string) => void }) {
  const [mode, setMode] = useState<"totp" | "recovery">("totp");
  const [code, setCode] = useState("");
  const [invalid, setInvalid] = useState(false);
  const [pending, setPending] = useState(false);
  const otpRef = useRef<HTMLDivElement>(null);
  const t = useT();

  useEffect(() => {
    if (mode === "totp") otpRef.current?.querySelector("input")?.focus();
  }, [mode]);

  const verify = (value: string) => {
    if (value.trim() === "" || pending) return;
    setPending(true);
    client.auth
      .signInTwoFactor({ code: value.trim() }, { headers: { Authorization: `Token ${token}` }, quietStatuses: [403] })
      .then((res) => onSuccess(res.data.data.token))
      .catch((error: unknown) => {
        if (axios.isAxiosError(error) && error.response?.status === 410) {
          onBack();
          return;
        }
        setInvalid(true);
        setCode("");
        if (mode === "totp") otpRef.current?.querySelector("input")?.focus();
      })
      .finally(() => setPending(false));
  };

  return (
    <form
      className="flex flex-col gap-5"
      onSubmit={(event) => {
        event.preventDefault();
        verify(code);
      }}
      noValidate
    >
      {mode === "totp" ? (
        <div className="flex flex-col gap-1.5">
          <Label>{t("Authentication code")}</Label>
          <InputOtp
            ref={otpRef}
            value={code}
            onChange={(value) => {
              setCode(value);
              setInvalid(false);
            }}
            onComplete={verify}
            isInvalid={invalid}
            isDisabled={pending}
            groupEvery={3}
            aria-label={t("Authentication code")}
          />
          <HintText isInvalid={invalid}>
            {invalid ? t("That code didn't work. Try the current one.") : t("Open your authenticator app to view the code.")}
          </HintText>
        </div>
      ) : (
        <Input
          label={t("Recovery code")}
          placeholder="xxxxx-xxxxx"
          value={code}
          onChange={(value) => {
            setCode(value);
            setInvalid(false);
          }}
          isInvalid={invalid}
          errorMessage={t("That recovery code didn't work.")}
          hint={t("Each recovery code can be used once.")}
          autoComplete="off"
          autoFocus
          inputDir="ltr"
          inputClassName="font-mono"
        />
      )}

      <Button type="submit" className="w-full" disabled={pending || code.trim() === ""} aria-busy={pending || undefined} leadingIcon={pending ? PendingGlyph : undefined}>
        {t("Verify")}
      </Button>

      <div className="flex items-center justify-between gap-3">
        <LinkButton size="small" variant="secondary" leadingIcon={RiArrowLeftLine} onClick={onBack}>
          {t("Back")}
        </LinkButton>
        <LinkButton
          size="small"
          onClick={() => {
            setMode(mode === "totp" ? "recovery" : "totp");
            setCode("");
            setInvalid(false);
          }}
        >
          {mode === "totp" ? t("Use a recovery code") : t("Use authenticator app")}
        </LinkButton>
      </div>
    </form>
  );
}

export default function SignIn() {
  const navigate = useNavigate();
  const location = useLocation();
  const isAuthenticated = useIsAuthenticated();
  const [pendingToken, setPendingToken] = useState<string | null>(null);
  const t = useT();

  const from = (location.state as { from?: string } | null)?.from ?? "/dashboard";

  if (isAuthenticated) {
    return <Navigate to={from} replace />;
  }

  const onSuccess = (token: string) => {
    authStore.setToken(token);
    navigate(from, { replace: true });
  };

  return (
    <div className="relative flex min-h-dvh flex-col items-center justify-center bg-background-full px-4 py-10">
      <div className="absolute top-4 end-4 flex items-center gap-2">
        <LanguageSwitch />
        <ThemeToggle appearance="segmented" />
      </div>

      <div className="flex w-full max-w-[400px] flex-col items-center gap-6">
        <section className="w-full rounded-3xl border border-border-button-default bg-background-full p-6 shadow-sm sm:p-8">
          <div className="mb-6 flex flex-col gap-4">
            <Logo className="size-11" />
            <h1 className="text-title-2-medium text-text-primary">{pendingToken ? t("Two-factor authentication") : t("Sign in to Elaina")}</h1>
          </div>

          {pendingToken ? (
            <TwoFactorStep token={pendingToken} onBack={() => setPendingToken(null)} onSuccess={onSuccess} />
          ) : (
            <PasswordStep onPending={setPendingToken} onSuccess={onSuccess} />
          )}
        </section>

        <footer className="flex items-center gap-2 text-caption-1-medium text-text-tertiary">
          <span>© 2021-{new Date().getFullYear()} Elaina</span>
          <span aria-hidden>·</span>
          <a
            href="https://github.com/wuhan005/Elaina"
            target="_blank"
            rel="noreferrer"
            className="flex items-center gap-1 rounded-md outline-none transition-colors hover:text-text-secondary focus-visible:ring-2 focus-visible:ring-border-focus-ring"
          >
            <RiGithubFill className="size-4" aria-hidden />
            GitHub
          </a>
        </footer>
      </div>
    </div>
  );
}
