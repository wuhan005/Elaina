import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { RiAddLine, RiDeleteBin6Line, RiEditLine, RiGlobalLine, RiProhibited2Line, RiStackLine } from "@remixicon/react";
import type { Template } from "@/api/api";
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
import { LANGUAGES_MAP } from "@/const/template";
import { translate, useT } from "@/i18n";
import { formatDateTime } from "@/utils/format";

const PAGE_SIZE = 20;

export function LanguageChips({ languages }: { languages?: string[] | null }) {
  const t = useT();
  if (!languages || languages.length === 0) {
    return <span className="text-body-medium text-text-tertiary">{t("None")}</span>;
  }
  return (
    <div className="flex flex-wrap gap-1">
      {languages.map((value) => (
        <Chip key={value} variant="caption" color={LANGUAGES_MAP[value]?.color ?? "neutral"}>
          {LANGUAGES_MAP[value]?.label ?? value}
        </Chip>
      ))}
    </div>
  );
}

export function NetworkChip({ allowed }: { allowed: boolean }) {
  const t = useT();
  return allowed ? (
    <Chip variant="caption" color="lime" className="gap-1">
      <RiGlobalLine className="size-3.5" aria-hidden />
      {t("Allowed")}
    </Chip>
  ) : (
    <Chip variant="caption" color="neutral" className="gap-1">
      <RiProhibited2Line className="size-3.5" aria-hidden />
      {t("Denied")}
    </Chip>
  );
}

export function formatTimeout(timeout: number): string {
  return timeout === 0 ? translate("10s (default)") : `${timeout}s`;
}

export function formatMemory(maxMemory: number): string {
  return maxMemory === 0 ? translate("Server ceiling") : `${maxMemory} MiB`;
}

export default function Templates() {
  const navigate = useNavigate();
  const t = useT();
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [data, setData] = useState<Template[]>([]);
  const [loading, setLoading] = useState(true);
  // Keep deleteTarget while the dialog closes so the name stays during the exit animation.
  const [deleteTarget, setDeleteTarget] = useState<Template | null>(null);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const load = useCallback(() => {
    setLoading(true);
    client.templates
      .listTemplates({ page, pageSize: PAGE_SIZE })
      .then((res) => {
        const { templates, total } = res.data.data;
        // Step back a page after deleting the last item on the last page.
        if (templates.length === 0 && page > 1) {
          setPage(page - 1);
          return;
        }
        setData(templates);
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
    const res = await client.templates.deleteTemplate({ id: deleteTarget.id });
    toast.success(res.data.data);
    load();
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);
  const showEmpty = !loading && data.length === 0;

  return (
    <>
      <PageHeader
        title={t("Templates")}
        actions={
          <Button leadingIcon={RiAddLine} onClick={() => navigate("/template/new")}>
            {t("New template")}
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
                {t(total === 1 ? "{count} template" : "{count} templates", { count: total.toLocaleString() })}
              </p>
            )}
          </div>
        </div>

        {showEmpty ? (
          <div className="border-t border-separator-border">
            <EmptyState
              icon={<RiStackLine className="size-6" aria-hidden />}
              title={t("No templates")}
              description={t("A sandbox needs a template for its languages and limits.")}
              action={
                <Button leadingIcon={RiAddLine} onClick={() => navigate("/template/new")}>
                  {t("New template")}
                </Button>
              }
            />
          </div>
        ) : (
          <div className="mt-2">
            <Table aria-label={t("Templates")} className="min-w-[900px]">
              <TableHeader>
                <TableColumn isRowHeader className="w-[220px]">{t("Name")}</TableColumn>
                <TableColumn>{t("Languages")}</TableColumn>
                <TableColumn className="w-[130px]">{t("Timeout")}</TableColumn>
                <TableColumn className="w-[140px]">{t("Memory")}</TableColumn>
                <TableColumn className="w-[120px]">{t("Network")}</TableColumn>
                <TableColumn className="w-[180px]">{t("Created at")}</TableColumn>
                <TableColumn className="w-[100px]">
                  <span className="sr-only">{t("Actions")}</span>
                </TableColumn>
              </TableHeader>
              <TableBody>
                {loading && data.length === 0
                  ? Array.from({ length: 5 }, (_, index) => (
                      <TableRow key={`skeleton-${index}`} id={`skeleton-${index}`}>
                        <TableCell><Skeleton shape="text" className="h-4 w-32" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-40" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-16" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-20" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-16" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-32" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-16" /></TableCell>
                      </TableRow>
                    ))
                  : data.map((template) => (
                      <TableRow key={template.id} id={template.id} className="hover:bg-background-secondary-default">
                        <TableCell>
                          <div className="flex min-w-0 flex-col">
                            <span className="truncate text-body-medium text-text-primary">{template.name}</span>
                            <span className="text-caption-1-medium text-text-tertiary">#{template.id}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <LanguageChips languages={template.language} />
                        </TableCell>
                        <TableCell>
                          <span className="whitespace-nowrap tabular-nums">{formatTimeout(template.timeout)}</span>
                        </TableCell>
                        <TableCell>
                          <span className="whitespace-nowrap tabular-nums">{formatMemory(template.maxMemory)}</span>
                        </TableCell>
                        <TableCell>
                          <NetworkChip allowed={template.internetAccess} />
                        </TableCell>
                        <TableCell>
                          <span className="whitespace-nowrap text-text-secondary tabular-nums">{formatDateTime(template.createdAt)}</span>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center justify-end gap-2">
                            <RowAction icon={RiEditLine} label={t("Edit")} onClick={() => navigate(`/template/${template.id}`)} />
                            <RowAction icon={RiDeleteBin6Line} label={t("Delete")} danger onClick={() => {
                                setDeleteTarget(template);
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
        description={t("Sandboxes using {name} will stop running code. This can't be undone.", { name: deleteTarget?.name ?? "" })}
        confirmLabel={t("Delete")}
        pendingLabel={t("Deleting…")}
        onConfirm={onDelete}
      />
    </>
  );
}
