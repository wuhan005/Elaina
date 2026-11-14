import { useState, type FormEvent } from "react";
import { Navigate, useLocation, useNavigate } from "react-router";
import { RiEyeLine, RiEyeOffLine, RiGithubFill, RiLockPasswordLine } from "@remixicon/react";
import { client } from "@/api/client";
import { ElainaMark } from "@/components/application/layout/elaina-mark";
import { ThemeToggle } from "@/components/application/theme/theme-toggle";
import { Button } from "@/components/base/buttons/button";
import { HintText } from "@/components/base/input/hint-text";
import { InputBase, TextField } from "@/components/base/input/input";
import { Label } from "@/components/base/input/label";
import { SpinnerGlyph } from "@/components/base/spinner/spinner";
import { toast } from "@/components/base/toast/toast";
import { authStore, useIsAuthenticated } from "@/store/auth";

const PendingGlyph = ({ className }: { className?: string }) => <SpinnerGlyph size="sm" className={className} />;

// 后端按 UTF-8 字节数校验密码长度（8-72）。
function validatePassword(password: string): string | null {
  const bytes = new TextEncoder().encode(password).length;
  if (bytes === 0) return "Password is required.";
  if (bytes < 8 || bytes > 72) return "Password must contain 8-72 UTF-8 bytes.";
  return null;
}

export default function SignIn() {
  const navigate = useNavigate();
  const location = useLocation();
  const isAuthenticated = useIsAuthenticated();
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  const from = (location.state as { from?: string } | null)?.from ?? "/dashboard";

  if (isAuthenticated && !pending) {
    return <Navigate to={from} replace />;
  }

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const message = validatePassword(password);
    setError(message);
    if (message) return;

    setPending(true);
    client.auth
      .signIn({ password })
      .then((res) => {
        authStore.setToken(res.data.data);
        toast.success("Signed in successfully");
        navigate(from, { replace: true });
      })
      .catch(() => undefined)
      .finally(() => setPending(false));
  };

  return (
    <div className="relative flex min-h-dvh flex-col items-center justify-center bg-background-full px-4 py-10">
      <div className="absolute top-4 end-4">
        <ThemeToggle appearance="segmented" />
      </div>

      <div className="flex w-full max-w-[400px] flex-col items-center gap-6">
        <section className="w-full rounded-3xl border border-border-button-default bg-background-full p-6 shadow-sm sm:p-8">
          <div className="mb-6 flex flex-col gap-5">
            <ElainaMark size="lg" />
            <div className="flex flex-col gap-1.5">
              <h1 className="text-title-2-medium text-text-primary">Welcome back</h1>
              <p className="text-body-regular text-text-secondary">
                Sign in with the administrator password to manage templates and sandboxes.
              </p>
            </div>
          </div>

          <form className="flex flex-col gap-6" onSubmit={onSubmit} noValidate>
            <TextField
              name="password"
              type={showPassword ? "text" : "password"}
              value={password}
              onChange={(value) => {
                setPassword(value);
                if (error) setError(null);
              }}
              isInvalid={error !== null}
              isRequired
              autoFocus
              autoComplete="current-password"
            >
              <Label isRequired>Password</Label>
              <InputBase
                placeholder="Enter your password"
                leadingIcon={RiLockPasswordLine}
                trailingAddon={
                  <button
                    type="button"
                    aria-label={showPassword ? "Hide password" : "Show password"}
                    onClick={() => setShowPassword((value) => !value)}
                    className="flex shrink-0 cursor-pointer items-center rounded-md text-foreground-icon-tertiary outline-none transition-colors hover:text-foreground-icon-primary focus-visible:ring-2 focus-visible:ring-border-focus-ring"
                  >
                    {showPassword ? <RiEyeOffLine className="size-5" aria-hidden /> : <RiEyeLine className="size-5" aria-hidden />}
                  </button>
                }
              />
              {error ? <HintText isInvalid>{error}</HintText> : null}
            </TextField>

            <Button
              type="submit"
              className="w-full"
              disabled={pending}
              aria-busy={pending || undefined}
              leadingIcon={pending ? PendingGlyph : undefined}
            >
              {pending ? "Signing in…" : "Sign in"}
            </Button>
          </form>
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
