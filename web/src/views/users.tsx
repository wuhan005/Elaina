import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useNavigate } from "react-router";
import { RiAddLine, RiDeleteBin6Line, RiKey2Line, RiShieldLine } from "@remixicon/react";
import type { User } from "@/api/api";
import { client } from "@/api/client";
import { PasswordField, validateNewPassword } from "@/components/application/auth/password-field";
import { PageHeader } from "@/components/application/layout/page-header";
import { RowAction } from "@/components/application/layout/row-action";
import { Avatar } from "@/components/base/avatar/avatar";
import { Chip } from "@/components/base/badges/chip";
import { Button } from "@/components/base/buttons/button";
import { Input } from "@/components/base/input/input";
import { AlertDialog, Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/base/modal/modal";
import { Skeleton } from "@/components/base/skeleton/skeleton";
import { SpinnerGlyph } from "@/components/base/spinner/spinner";
import { Table, TableBody, TableCell, TableColumn, TableHeader, TableRow } from "@/components/base/table/table";
import { toast } from "@/components/base/toast/toast";
import { useT } from "@/i18n";
import { useCurrentUser } from "@/store/auth";
import { formatDateTime } from "@/utils/format";

const PendingGlyph = ({ className }: { className?: string }) => <SpinnerGlyph size="sm" className={className} />;

const USERNAME_PATTERN = /^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,31}$/;

function CreateUserModal({ isOpen, onOpenChange, onCreated }: { isOpen: boolean; onOpenChange: (open: boolean) => void; onCreated: () => void }) {
  const t = useT();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<{ username?: string; password?: string }>({});
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    setUsername("");
    setPassword("");
    setErrors({});
  }, [isOpen]);

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const next = {
      username: USERNAME_PATTERN.test(username) ? undefined : t("2-32 letters, digits, '_', '.' or '-'."),
      password: validateNewPassword(password),
    };
    setErrors(next);
    if (next.username || next.password) return;

    setPending(true);
    client.users
      .createUser({ username, password })
      .then((res) => {
        toast.success(t("Created {username}", { username: res.data.data.username }));
        onOpenChange(false);
        onCreated();
      })
      .catch(() => undefined)
      .finally(() => setPending(false));
  };

  return (
    <Modal isOpen={isOpen} onOpenChange={onOpenChange} size="sm">
      {({ close }) => (
        <form className="contents" onSubmit={onSubmit} noValidate>
          <ModalHeader title={t("New user")} />
          <ModalBody className="flex flex-col gap-4">
            <Input
              label={t("Username")}
              value={username}
              onChange={setUsername}
              isRequired
              isInvalid={!!errors.username}
              errorMessage={errors.username}
              autoComplete="off"
              autoFocus
              inputDir="ltr"
            />
            <PasswordField label={t("Password")} value={password} onChange={setPassword} error={errors.password} hint={t("8-72 characters.")} autoComplete="new-password" />
          </ModalBody>
          <ModalFooter>
            <Button variant="secondary" onClick={close}>
              {t("Cancel")}
            </Button>
            <Button type="submit" disabled={pending} leadingIcon={pending ? PendingGlyph : undefined}>
              {t("Create")}
            </Button>
          </ModalFooter>
        </form>
      )}
    </Modal>
  );
}

function ResetPasswordModal({ user, isOpen, onOpenChange }: { user: User | null; isOpen: boolean; onOpenChange: (open: boolean) => void }) {
  const t = useT();
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string>();
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    setPassword("");
    setError(undefined);
  }, [isOpen]);

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const message = validateNewPassword(password);
    setError(message);
    if (message || !user) return;

    setPending(true);
    client.users
      .resetUserPassword({ id: user.id }, { password })
      .then((res) => {
        toast.success(res.data.data, { description: t("{username} has been signed out.", { username: user.username }) });
        onOpenChange(false);
      })
      .catch(() => undefined)
      .finally(() => setPending(false));
  };

  return (
    <Modal isOpen={isOpen} onOpenChange={onOpenChange} size="sm">
      {({ close }) => (
        <form className="contents" onSubmit={onSubmit} noValidate>
          <ModalHeader title={t("Set password for {username}", { username: user?.username ?? "" })} description={t("They will be signed out of all sessions.")} />
          <ModalBody>
            <PasswordField label={t("New password")} value={password} onChange={setPassword} error={error} hint={t("8-72 characters.")} autoComplete="new-password" autoFocus />
          </ModalBody>
          <ModalFooter>
            <Button variant="secondary" onClick={close}>
              {t("Cancel")}
            </Button>
            <Button type="submit" disabled={pending} leadingIcon={pending ? PendingGlyph : undefined}>
              {t("Set password")}
            </Button>
          </ModalFooter>
        </form>
      )}
    </Modal>
  );
}

type Action = "password" | "2fa" | "delete";

export default function Users() {
  const t = useT();
  const navigate = useNavigate();
  const currentUser = useCurrentUser();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  // Keep target while a dialog closes so its content stays during the exit animation.
  const [target, setTarget] = useState<User | null>(null);
  const [action, setAction] = useState<Action | null>(null);

  const load = useCallback(() => {
    client.users
      .listUsers()
      .then((res) => setUsers(res.data.data))
      .catch(() => undefined)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const open = (user: User, next: Action) => {
    setTarget(user);
    setAction(next);
  };
  const setOpen = (kind: Action) => (isOpen: boolean) => setAction(isOpen ? kind : null);

  return (
    <>
      <PageHeader
        title={t("Users")}
        actions={
          <Button leadingIcon={RiAddLine} onClick={() => setCreateOpen(true)}>
            {t("New user")}
          </Button>
        }
      />

      <section className="flex w-full flex-col rounded-2xl border border-border-table bg-background-full pt-2 pb-2">
        <div className="px-4 py-2">
          {loading ? (
            <Skeleton shape="text" className="my-1 h-3.5 w-20" />
          ) : (
            <p className="text-body-medium text-text-secondary tabular-nums">
              {t(users.length === 1 ? "{count} user" : "{count} users", { count: users.length })}
            </p>
          )}
        </div>
        <Table aria-label={t("Users")} className="mt-2 min-w-[720px]">
          <TableHeader>
            <TableColumn isRowHeader>{t("Username")}</TableColumn>
            <TableColumn className="w-[150px]">{t("Two-factor")}</TableColumn>
            <TableColumn className="w-[190px]">{t("Last sign-in")}</TableColumn>
            <TableColumn className="w-[190px]">{t("Created")}</TableColumn>
            <TableColumn className="w-[140px]">
              <span className="sr-only">{t("Actions")}</span>
            </TableColumn>
          </TableHeader>
          <TableBody>
            {loading
              ? Array.from({ length: 3 }, (_, index) => (
                  <TableRow key={`skeleton-${index}`} id={`skeleton-${index}`}>
                    <TableCell><Skeleton shape="text" className="h-4 w-32" /></TableCell>
                    <TableCell><Skeleton shape="text" className="h-4 w-12" /></TableCell>
                    <TableCell><Skeleton shape="text" className="h-4 w-32" /></TableCell>
                    <TableCell><Skeleton shape="text" className="h-4 w-32" /></TableCell>
                    <TableCell><Skeleton shape="text" className="h-4 w-16" /></TableCell>
                  </TableRow>
                ))
              : users.map((user) => {
                  const isSelf = user.id === currentUser?.id;
                  return (
                    <TableRow key={user.id} id={user.id} className="hover:bg-background-secondary-default">
                      <TableCell>
                        <div className="flex min-w-0 items-center gap-2.5">
                          <Avatar size="sm" color="neutral" initials={user.username.slice(0, 1).toUpperCase()} />
                          <span className="truncate text-body-medium text-text-primary">{user.username}</span>
                          {isSelf && (
                            <Chip variant="caption" color="blue">
                              {t("You")}
                            </Chip>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        {user.twoFactorEnabled ? (
                          <Chip variant="caption" color="lime">{t("On")}</Chip>
                        ) : (
                          <Chip variant="caption" color="neutral">{t("Off")}</Chip>
                        )}
                      </TableCell>
                      <TableCell>
                        <span className="whitespace-nowrap text-text-secondary tabular-nums">
                          {user.lastSignInAt ? formatDateTime(user.lastSignInAt) : t("Never")}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="whitespace-nowrap text-text-secondary tabular-nums">{formatDateTime(user.createdAt)}</span>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center justify-end gap-2">
                          {isSelf ? (
                            <Button variant="secondary" size="small" onClick={() => navigate("/account")}>
                              {t("Settings")}
                            </Button>
                          ) : (
                            <>
                              <RowAction icon={RiKey2Line} label={t("Set password")} onClick={() => open(user, "password")} />
                              {user.twoFactorEnabled && (
                                <RowAction icon={RiShieldLine} label={t("Turn off two-factor")} onClick={() => open(user, "2fa")} />
                              )}
                              <RowAction icon={RiDeleteBin6Line} label={t("Delete")} danger onClick={() => open(user, "delete")} />
                            </>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
          </TableBody>
        </Table>
      </section>

      <CreateUserModal isOpen={createOpen} onOpenChange={setCreateOpen} onCreated={load} />
      <ResetPasswordModal user={target} isOpen={action === "password"} onOpenChange={setOpen("password")} />

      <AlertDialog
        isOpen={action === "2fa"}
        onOpenChange={setOpen("2fa")}
        title={t("Turn off two-factor for {username}?", { username: target?.username ?? "" })}
        description={t("Use this when they've lost both their authenticator and recovery codes. They will be signed out and can sign in with just their password.")}
        confirmLabel={t("Turn off")}
        onConfirm={async () => {
          if (!target) return;
          const res = await client.users.resetUserTwoFactor({ id: target.id });
          toast.success(res.data.data);
          load();
        }}
      />

      <AlertDialog
        isOpen={action === "delete"}
        onOpenChange={setOpen("delete")}
        title={t("Delete {name}?", { name: target?.username ?? "" })}
        description={t("They will be signed out immediately. This can't be undone.")}
        confirmLabel={t("Delete")}
        onConfirm={async () => {
          if (!target) return;
          const res = await client.users.deleteUser({ id: target.id });
          toast.success(res.data.data);
          load();
        }}
      />
    </>
  );
}
