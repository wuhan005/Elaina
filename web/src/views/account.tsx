import { useEffect, useRef, useState, type FormEvent, type ReactNode } from "react";
import { RiFileCopyLine, RiShieldCheckLine } from "@remixicon/react";
import type { TwoFactorSetupResp, User } from "@/api/api";
import { client } from "@/api/client";
import { PasswordField, validateNewPassword } from "@/components/application/auth/password-field";
import { RecoveryCodes } from "@/components/application/auth/recovery-codes";
import { refreshCurrentUser } from "@/components/application/layout/app-layout";
import { PageHeader } from "@/components/application/layout/page-header";
import { FormSection, Panel } from "@/components/application/layout/panel";
import { Chip } from "@/components/base/badges/chip";
import { Button } from "@/components/base/buttons/button";
import { IconButton } from "@/components/base/buttons/icon-button";
import { HintText } from "@/components/base/input/hint-text";
import { InputOtp } from "@/components/base/input-otp/input-otp";
import { Label } from "@/components/base/input/label";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/base/modal/modal";
import { Skeleton } from "@/components/base/skeleton/skeleton";
import { Spinner, SpinnerGlyph } from "@/components/base/spinner/spinner";
import { toast } from "@/components/base/toast/toast";
import { useCurrentUser } from "@/store/auth";
import { formatDateTime } from "@/utils/format";

const PendingGlyph = ({ className }: { className?: string }) => <SpinnerGlyph size="sm" className={className} />;

function ChangePasswordForm() {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [errors, setErrors] = useState<{ current?: string; next?: string; confirm?: string }>({});
  const [saving, setSaving] = useState(false);

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const nextErrors = {
      current: current === "" ? "Enter your current password." : undefined,
      next: validateNewPassword(next),
      confirm: confirm !== next ? "Passwords don't match." : undefined,
    };
    setErrors(nextErrors);
    if (nextErrors.current || nextErrors.next || nextErrors.confirm) return;

    setSaving(true);
    client.account
      .changePassword({ currentPassword: current, newPassword: next })
      .then((res) => {
        toast.success(res.data.data, { description: "Other sessions have been signed out." });
        setCurrent("");
        setNext("");
        setConfirm("");
      })
      .catch(() => undefined)
      .finally(() => setSaving(false));
  };

  return (
    <form onSubmit={onSubmit} noValidate className="flex max-w-md flex-col gap-4">
      <PasswordField label="Current password" value={current} onChange={setCurrent} error={errors.current} autoComplete="current-password" />
      <PasswordField label="New password" value={next} onChange={setNext} error={errors.next} hint="8-72 characters." autoComplete="new-password" />
      <PasswordField label="Confirm new password" value={confirm} onChange={setConfirm} error={errors.confirm} autoComplete="new-password" />
      <div>
        <Button type="submit" disabled={saving} aria-busy={saving || undefined} leadingIcon={saving ? PendingGlyph : undefined}>
          Update password
        </Button>
      </div>
    </form>
  );
}

function SetupTwoFactorModal({ isOpen, onOpenChange, username }: { isOpen: boolean; onOpenChange: (open: boolean) => void; username: string }) {
  const [setup, setSetup] = useState<TwoFactorSetupResp | null>(null);
  const [code, setCode] = useState("");
  const [invalid, setInvalid] = useState(false);
  const [pending, setPending] = useState(false);
  const [codes, setCodes] = useState<string[] | null>(null);
  const otpRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isOpen) return;
    setSetup(null);
    setCodes(null);
    setCode("");
    setInvalid(false);
    client.account
      .setupTwoFactor()
      .then((res) => setSetup(res.data.data))
      .catch(() => onOpenChange(false));
  }, [isOpen, onOpenChange]);

  const enable = (value: string) => {
    if (value.length !== 6 || pending) return;
    setPending(true);
    client.account
      .enableTwoFactor({ code: value }, { quietStatuses: [403] })
      .then((res) => {
        setCodes(res.data.data.codes);
        void refreshCurrentUser();
      })
      .catch(() => {
        setInvalid(true);
        setCode("");
        otpRef.current?.querySelector("input")?.focus();
      })
      .finally(() => setPending(false));
  };

  return (
    <Modal isOpen={isOpen} onOpenChange={onOpenChange} isDismissable={!codes} size="md">
      {({ close }) =>
        codes ? (
          <>
            <ModalHeader
              title="Save your recovery codes"
              description="If you lose your phone, each code lets you sign in once. They won't be shown again."
              showCloseButton={false}
            />
            <ModalBody>
              <RecoveryCodes codes={codes} username={username} />
            </ModalBody>
            <ModalFooter>
              <Button onClick={close}>Done</Button>
            </ModalFooter>
          </>
        ) : (
          <>
            <ModalHeader title="Set up two-factor authentication" />
            <ModalBody className="flex flex-col gap-5">
              {setup ? (
                <>
                  <div className="flex flex-col gap-4 sm:flex-row sm:items-start">
                    <img src={setup.qrCode} alt="QR code for your authenticator app" className="size-36 shrink-0 rounded-xl border border-border-button-default bg-white p-1.5" />
                    <div className="flex min-w-0 flex-col gap-2">
                      <p className="text-body-regular text-text-secondary">
                        Scan the QR code with an authenticator app such as Google Authenticator, 1Password or Authy. If you can't scan it, enter this key instead:
                      </p>
                      <div className="flex items-center gap-2">
                        <code dir="ltr" className="min-w-0 flex-1 break-all rounded-lg bg-background-secondary-default px-2 py-1.5 font-mono text-caption-1-medium text-text-primary">
                          {setup.secret}
                        </code>
                        <IconButton
                          icon={RiFileCopyLine}
                          size="small"
                          aria-label="Copy key"
                          onClick={() => navigator.clipboard.writeText(setup.secret).then(() => toast.success("Key copied"))}
                        />
                      </div>
                    </div>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label>Code from the app</Label>
                    <InputOtp
                      ref={otpRef}
                      value={code}
                      onChange={(value) => {
                        setCode(value);
                        setInvalid(false);
                      }}
                      onComplete={enable}
                      isInvalid={invalid}
                      isDisabled={pending}
                      groupEvery={3}
                      aria-label="Code from the app"
                    />
                    {invalid ? <HintText isInvalid>That code didn't work. Try the current one.</HintText> : null}
                  </div>
                </>
              ) : (
                <div className="flex h-48 items-center justify-center">
                  <Spinner size="lg" />
                </div>
              )}
            </ModalBody>
            <ModalFooter>
              <Button variant="secondary" onClick={close}>
                Cancel
              </Button>
              <Button onClick={() => enable(code)} disabled={!setup || code.length !== 6 || pending} leadingIcon={pending ? PendingGlyph : undefined}>
                Turn on
              </Button>
            </ModalFooter>
          </>
        )
      }
    </Modal>
  );
}

// PasswordConfirmModal asks for the password again before sensitive actions such as turning off two-factor authentication.
function PasswordConfirmModal({
  isOpen,
  onOpenChange,
  title,
  description,
  confirmLabel,
  danger,
  onConfirm,
  children,
}: {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  confirmLabel: string;
  danger?: boolean;
  onConfirm: (password: string) => Promise<unknown>;
  children?: ReactNode;
}) {
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string>();
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setPassword("");
      setError(undefined);
    }
  }, [isOpen]);

  return (
    <Modal isOpen={isOpen} onOpenChange={onOpenChange} size="sm">
      {({ close }) =>
        children ?? (
          <form
            className="contents"
            onSubmit={(event) => {
              event.preventDefault();
              if (password === "") {
                setError("Enter your password.");
                return;
              }
              setPending(true);
              onConfirm(password)
                .catch(() => setPassword(""))
                .finally(() => setPending(false));
            }}
          >
            <ModalHeader title={title} description={description} />
            <ModalBody>
              <PasswordField label="Password" value={password} onChange={setPassword} error={error} autoComplete="current-password" autoFocus />
            </ModalBody>
            <ModalFooter>
              <Button variant="secondary" onClick={close}>
                Cancel
              </Button>
              <Button type="submit" variant={danger ? "danger" : "primary"} disabled={pending} leadingIcon={pending ? PendingGlyph : undefined}>
                {confirmLabel}
              </Button>
            </ModalFooter>
          </form>
        )
      }
    </Modal>
  );
}

function TwoFactorSection({ user }: { user: User }) {
  const [setupOpen, setSetupOpen] = useState(false);
  const [disableOpen, setDisableOpen] = useState(false);
  const [regenerateOpen, setRegenerateOpen] = useState(false);
  const [newCodes, setNewCodes] = useState<string[] | null>(null);

  return (
    <>
      <div className="flex flex-col gap-4 rounded-2xl border border-border-button-default p-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 items-start gap-3">
          <span className="flex size-10 shrink-0 items-center justify-center rounded-2lg bg-background-tertiary-default text-foreground-icon-primary">
            <RiShieldCheckLine className="size-5" aria-hidden />
          </span>
          <div className="flex min-w-0 flex-col gap-0.5">
            <div className="flex items-center gap-2">
              <span className="text-body-medium text-text-primary">Authenticator app</span>
              {user.twoFactorEnabled ? (
                <Chip variant="caption" color="lime">On</Chip>
              ) : (
                <Chip variant="caption" color="neutral">Off</Chip>
              )}
            </div>
            <span className="text-body-regular text-text-secondary">
              {user.twoFactorEnabled
                ? `${user.recoveryCodesRemaining} of 10 recovery codes left.`
                : "Ask for a 6-digit code from an app after the password."}
            </span>
          </div>
        </div>
        <div className="flex shrink-0 flex-wrap gap-2">
          {user.twoFactorEnabled ? (
            <>
              <Button variant="secondary" size="small" onClick={() => setRegenerateOpen(true)}>
                New recovery codes
              </Button>
              <Button variant="secondary" size="small" onClick={() => setDisableOpen(true)}>
                Turn off
              </Button>
            </>
          ) : (
            <Button size="small" onClick={() => setSetupOpen(true)}>
              Set up
            </Button>
          )}
        </div>
      </div>

      <SetupTwoFactorModal isOpen={setupOpen} onOpenChange={setSetupOpen} username={user.username} />

      <PasswordConfirmModal
        isOpen={disableOpen}
        onOpenChange={setDisableOpen}
        title="Turn off two-factor authentication?"
        description="Signing in will only need your password. Your recovery codes stop working."
        confirmLabel="Turn off"
        danger
        onConfirm={(password) =>
          client.account.disableTwoFactor({ password }).then((res) => {
            toast.success(res.data.data);
            setDisableOpen(false);
            return refreshCurrentUser();
          })
        }
      />

      <PasswordConfirmModal
        isOpen={regenerateOpen}
        onOpenChange={(open) => {
          setRegenerateOpen(open);
          if (!open) setNewCodes(null);
        }}
        title="Generate new recovery codes"
        description="Your current recovery codes will stop working."
        confirmLabel="Generate"
        onConfirm={(password) =>
          client.account.regenerateRecoveryCodes({ password }).then((res) => {
            setNewCodes(res.data.data.codes);
            return refreshCurrentUser();
          })
        }
      >
        {newCodes ? (
          <>
            <ModalHeader title="New recovery codes" description="Your old codes no longer work. These won't be shown again." showCloseButton={false} />
            <ModalBody>
              <RecoveryCodes codes={newCodes} username={user.username} />
            </ModalBody>
            <ModalFooter>
              <Button
                onClick={() => {
                  setRegenerateOpen(false);
                  setNewCodes(null);
                }}
              >
                Done
              </Button>
            </ModalFooter>
          </>
        ) : undefined}
      </PasswordConfirmModal>
    </>
  );
}

function Detail({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-caption-1-medium text-text-tertiary">{label}</dt>
      <dd className="text-body-medium text-text-primary">{children}</dd>
    </div>
  );
}

export default function Account() {
  const user = useCurrentUser();

  return (
    <>
      <PageHeader title="Account settings" />

      <Panel bodyClassName="p-0">
        <FormSection title="Profile">
          {user ? (
            <dl className="grid gap-4 sm:grid-cols-3">
              <Detail label="Username">{user.username}</Detail>
              <Detail label="Created">{formatDateTime(user.createdAt)}</Detail>
              <Detail label="Last sign-in">{user.lastSignInAt ? formatDateTime(user.lastSignInAt) : "—"}</Detail>
            </dl>
          ) : (
            <Skeleton shape="text" className="h-4 w-60" />
          )}
        </FormSection>

        <FormSection title="Password" description="Changing it signs you out everywhere else.">
          <ChangePasswordForm />
        </FormSection>

        <FormSection title="Two-factor authentication">
          {user ? <TwoFactorSection user={user} /> : <Skeleton className="h-20 w-full rounded-2xl" />}
        </FormSection>
      </Panel>
    </>
  );
}
