import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router";
import {
  RiAddLine,
  RiCodeBoxLine,
  RiDeleteBin6Line,
  RiEditLine,
  RiExternalLinkLine,
  RiFileCopyLine,
  RiHistoryLine,
  RiLockLine,
  RiPencilLine,
} from "@remixicon/react";
import type { Sandbox } from "@/api/api";
import { client } from "@/api/client";
import { PageHeader } from "@/components/application/layout/page-header";
import { EmptyState } from "@/components/application/layout/panel";
import { RowAction } from "@/components/application/layout/row-action";
import { Chip } from "@/components/base/badges/chip";
import { Button } from "@/components/base/buttons/button";
import { AlertDialog } from "@/components/base/modal/modal";
import { Pagination } from "@/components/base/pagination/pagination";
import { Skeleton } from "@/components/base/skeleton/skeleton";
import { Table, TableBody, TableCell, TableColumn, TableHeader, TableRow } from "@/components/base/table/table";
import { toast } from "@/components/base/toast/toast";
import { translate, useT } from "@/i18n";
import { formatDateTime } from "@/utils/format";

const PAGE_SIZE = 20;

export function runnerUrl(uid: string): string {
  return `${window.location.origin}/r/${uid}`;
}

export function copyRunnerUrl(uid: string) {
  navigator.clipboard
    .writeText(runnerUrl(uid))
    .then(() => toast.success(translate("Link copied")))
    .catch(() => toast.error(translate("Could not copy the link")));
}

export function EditableChip({ editable }: { editable: boolean }) {
  const t = useT();
  return editable ? (
    <Chip variant="caption" color="blue" className="gap-1">
      <RiPencilLine className="size-3.5" aria-hidden />
      {t("Editable")}
    </Chip>
  ) : (
    <Chip variant="caption" color="neutral" className="gap-1">
      <RiLockLine className="size-3.5" aria-hidden />
      {t("Read-only")}
    </Chip>
  );
}

export function TemplateName({ sandbox }: { sandbox: Sandbox }) {
  const t = useT();
  return sandbox.template ? (
    <span className="truncate text-body-medium text-text-primary">{sandbox.template.name}</span>
  ) : (
    <Chip variant="caption" color="rose">
      {t("Template unavailable")}
    </Chip>
  );
}

export function CodePreview({ code }: { code: string }) {
  const t = useT();
  const firstLine = code.split("\n").find((line) => line.trim() !== "") ?? "";
  if (firstLine === "") {
    return <span className="text-body-medium text-text-tertiary">{t("Empty")}</span>;
  }
  return (
    <code dir="ltr" className="block max-w-[240px] truncate rounded-md bg-background-secondary-default px-1.5 py-0.5 font-mono text-caption-1-medium text-text-secondary">
      {firstLine}
    </code>
  );
}

export default function Sandboxes() {
  const navigate = useNavigate();
  const t = useT();
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [data, setData] = useState<Sandbox[]>([]);
  const [loading, setLoading] = useState(true);
  // Keep deleteTarget while the dialog closes so the name stays during the exit animation.
  const [deleteTarget, setDeleteTarget] = useState<Sandbox | null>(null);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const load = useCallback(() => {
    setLoading(true);
    client.sandboxes
      .listSandboxes({ page, pageSize: PAGE_SIZE })
      .then((res) => {
        const { sandboxes, total } = res.data.data;
        if (sandboxes.length === 0 && page > 1) {
          setPage(page - 1);
          return;
        }
        setData(sandboxes);
        setTotal(total);
      })
      .catch(() => undefined)
      .finally(() => setLoading(false));
  }, [page]);

  useEffect(() => {
    load();
  }, [load]);

  const onDelete = async () => {
    if (!deleteTarget) return;
    const res = await client.sandboxes.deleteSandbox({ id: deleteTarget.id });
    toast.success(res.data.data);
    load();
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);
  const showEmpty = !loading && data.length === 0;

  return (
    <>
      <PageHeader
        title={t("Sandboxes")}
        actions={
          <Button leadingIcon={RiAddLine} onClick={() => navigate("/sandbox/new")}>
            {t("New sandbox")}
          </Button>
        }
      />

      <section className="flex w-full flex-col rounded-2xl border border-border-table bg-background-full pt-2">
        <div className="flex items-center justify-between gap-3 px-4 py-2">
          <div className="flex flex-col">
            {loading && data.length === 0 ? (
              <Skeleton shape="text" className="my-1 h-3.5 w-24" />
            ) : (
              <p className="text-body-medium text-text-secondary tabular-nums">
                {t(total === 1 ? "{count} sandbox" : "{count} sandboxes", { count: total.toLocaleString() })}
              </p>
            )}
          </div>
        </div>

        {showEmpty ? (
          <div className="border-t border-separator-border">
            <EmptyState
              icon={<RiCodeBoxLine className="size-6" aria-hidden />}
              title={t("No sandboxes")}
              description={t("Each sandbox gets a runner page at /r/:uid.")}
              action={
                <Button leadingIcon={RiAddLine} onClick={() => navigate("/sandbox/new")}>
                  {t("New sandbox")}
                </Button>
              }
            />
          </div>
        ) : (
          <div className="mt-2">
            <Table aria-label={t("Sandboxes")} className="min-w-[960px]">
              <TableHeader>
                <TableColumn isRowHeader className="w-[230px]">{t("Name")}</TableColumn>
                <TableColumn className="w-[180px]">{t("Template")}</TableColumn>
                <TableColumn className="w-[120px]">{t("Mode")}</TableColumn>
                <TableColumn>{t("Code")}</TableColumn>
                <TableColumn className="w-[180px]">{t("Created at")}</TableColumn>
                <TableColumn className="w-[210px]">
                  <span className="sr-only">{t("Actions")}</span>
                </TableColumn>
              </TableHeader>
              <TableBody>
                {loading && data.length === 0
                  ? Array.from({ length: 5 }, (_, index) => (
                      <TableRow key={`skeleton-${index}`} id={`skeleton-${index}`}>
                        <TableCell><Skeleton shape="text" className="h-4 w-36" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-24" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-16" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-48" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-32" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-24" /></TableCell>
                      </TableRow>
                    ))
                  : data.map((sandbox) => (
                      <TableRow key={sandbox.id} id={sandbox.id} className="hover:bg-background-secondary-default">
                        <TableCell>
                          <div className="flex min-w-0 flex-col">
                            <span className="truncate text-body-medium text-text-primary">{sandbox.name}</span>
                            <span dir="ltr" className="truncate font-mono text-caption-1-medium text-text-tertiary">{sandbox.uid}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <TemplateName sandbox={sandbox} />
                        </TableCell>
                        <TableCell>
                          <EditableChip editable={sandbox.editable} />
                        </TableCell>
                        <TableCell>
                          <CodePreview code={sandbox.placeholder} />
                        </TableCell>
                        <TableCell>
                          <span className="whitespace-nowrap text-text-secondary tabular-nums">{formatDateTime(sandbox.createdAt)}</span>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center justify-end gap-2">
                            <RowAction icon={RiExternalLinkLine} label={t("Open runner")} href={`/r/${sandbox.uid}`} />
                            <RowAction icon={RiFileCopyLine} label={t("Copy runner link")} onClick={() => copyRunnerUrl(sandbox.uid)} />
                            <RowAction icon={RiHistoryLine} label={t("Runs")} onClick={() => navigate(`/execution?sandbox=${sandbox.id}`)} />
                            <RowAction icon={RiEditLine} label={t("Edit")} onClick={() => navigate(`/sandbox/${sandbox.id}`)} />
                            <RowAction
                              icon={RiDeleteBin6Line}
                              label={t("Delete")}
                              danger
                              onClick={() => {
                                setDeleteTarget(sandbox);
                                setDeleteOpen(true);
                              }}
                            />
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
              </TableBody>
            </Table>
          </div>
        )}

        {totalPages > 1 ? (
          <div className="border-t border-separator-border px-4 py-3">
            <Pagination page={page} totalPages={totalPages} onChange={setPage} />
          </div>
        ) : (
          <div className="h-2" />
        )}
      </section>

      <AlertDialog
        isOpen={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={t("Delete {name}?", { name: deleteTarget?.name ?? "" })}
        description={t("Its runner link will stop working. This can't be undone.")}
        confirmLabel={t("Delete")}
        pendingLabel={t("Deleting…")}
        onConfirm={onDelete}
      />
    </>
  );
}
