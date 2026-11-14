import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router";
import { RiEyeLine, RiPulseLine } from "@remixicon/react";
import type { ExecutionSummary, Sandbox } from "@/api/api";
import { client } from "@/api/client";
import { ExecutionDrawer } from "@/components/application/executions/execution-drawer";
import { EXECUTION_STATUSES, ExecutionStatusChip, formatDuration, type ExecutionStatus } from "@/components/application/executions/execution-status";
import { PageHeader } from "@/components/application/layout/page-header";
import { EmptyState } from "@/components/application/layout/panel";
import { RowAction } from "@/components/application/layout/row-action";
import { Chip } from "@/components/base/badges/chip";
import { Pagination } from "@/components/base/pagination/pagination";
import { Select, SelectItem } from "@/components/base/select/select";
import { Skeleton } from "@/components/base/skeleton/skeleton";
import { Table, TableBody, TableCell, TableColumn, TableHeader, TableRow } from "@/components/base/table/table";
import { LANGUAGES_MAP } from "@/const/template";
import { useT } from "@/i18n";
import { formatDateTime } from "@/utils/format";

const PAGE_SIZE = 20;
const ALL = "all";

export function ExecutionSandboxName({ execution }: { execution: ExecutionSummary }) {
  const sandbox = execution.sandbox;
  if (!sandbox) return <span className="text-text-tertiary">#{execution.sandboxID}</span>;
  return (
    <div className="flex min-w-0 flex-col">
      <span className={sandbox.deleted ? "truncate text-body-medium text-text-tertiary line-through" : "truncate text-body-medium text-text-primary"}>
        {sandbox.name}
      </span>
      <span dir="ltr" className="truncate font-mono text-caption-1-medium text-text-tertiary">{sandbox.uid}</span>
    </div>
  );
}

export function LanguageChip({ language }: { language: string }) {
  if (!language) return <span className="text-text-tertiary">—</span>;
  const info = LANGUAGES_MAP[language];
  return (
    <Chip variant="caption" color={info?.color ?? "neutral"}>
      {info?.label ?? language}
    </Chip>
  );
}

export default function Executions() {
  const t = useT();
  const [params, setParams] = useSearchParams();
  const sandboxID = Number(params.get("sandbox")) || 0;
  const status = (params.get("status") ?? "") as ExecutionStatus | "";
  const page = Math.max(1, Number(params.get("page")) || 1);

  const [sandboxes, setSandboxes] = useState<Sandbox[]>([]);
  const [data, setData] = useState<ExecutionSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [openID, setOpenID] = useState<number | null>(null);

  const update = (next: { sandbox?: number; status?: string; page?: number }) => {
    const merged = new URLSearchParams(params);
    for (const [key, value] of Object.entries(next)) {
      if (!value || (key === "page" && value === 1)) merged.delete(key);
      else merged.set(key, String(value));
    }
    if (next.page === undefined) merged.delete("page");
    setParams(merged, { replace: true });
  };

  useEffect(() => {
    client.sandboxes
      .listSandboxes({ page: 1, pageSize: 500 })
      .then((res) => setSandboxes(res.data.data.sandboxes))
      .catch(() => undefined);
  }, []);

  const load = useCallback(() => {
    setLoading(true);
    client.executions
      .listExecutions({
        page,
        pageSize: PAGE_SIZE,
        sandboxID: sandboxID || undefined,
        status: status || undefined,
      })
      .then((res) => {
        setData(res.data.data.executions);
        setTotal(res.data.data.total);
      })
      .catch(() => undefined)
      .finally(() => setLoading(false));
  }, [page, sandboxID, status]);

  useEffect(() => {
    load();
  }, [load]);

  const closeDrawer = useCallback(() => setOpenID(null), []);
  const totalPages = Math.ceil(total / PAGE_SIZE);
  const filtered = sandboxID !== 0 || status !== "";

  return (
    <>
      <PageHeader title={t("Runs")} />

      <section className="flex w-full flex-col rounded-2xl border border-border-table bg-background-full pt-2">
        <div className="flex flex-col gap-3 px-4 py-2 sm:flex-row sm:items-center sm:justify-between">
          {loading && data.length === 0 ? (
            <Skeleton shape="text" className="my-1 h-3.5 w-20" />
          ) : (
            <p className="text-body-medium text-text-secondary tabular-nums">
              {t(total === 1 ? "{count} run" : "{count} runs", { count: total.toLocaleString() })}
            </p>
          )}
          <div className="flex flex-wrap items-center gap-2.5">
            <Select
              aria-label={t("Filter by sandbox")}
              placeholder={t("Sandbox")}
              popoverClassName="min-w-48"
              selectedKey={sandboxID ? String(sandboxID) : ALL}
              onSelectionChange={(key) => update({ sandbox: key === ALL ? 0 : Number(key) })}
            >
              {[
                <SelectItem key={ALL} id={ALL} textValue={t("All sandboxes")}>
                  {t("All sandboxes")}
                </SelectItem>,
                ...sandboxes.map((sandbox) => (
                  <SelectItem key={sandbox.id} id={String(sandbox.id)} textValue={sandbox.name}>
                    {sandbox.name}
                  </SelectItem>
                )),
              ]}
            </Select>
            <Select
              aria-label={t("Filter by status")}
              popoverClassName="min-w-40"
              selectedKey={status || ALL}
              onSelectionChange={(key) => update({ status: key === ALL ? "" : String(key) })}
            >
              {[
                <SelectItem key={ALL} id={ALL} textValue={t("All statuses")}>
                  {t("All statuses")}
                </SelectItem>,
                ...EXECUTION_STATUSES.map((option) => (
                  <SelectItem key={option.value} id={option.value} textValue={t(option.label)}>
                    {t(option.label)}
                  </SelectItem>
                )),
              ]}
            </Select>
          </div>
        </div>

        {!loading && data.length === 0 ? (
          <div className="border-t border-separator-border">
            <EmptyState
              icon={<RiPulseLine className="size-6" aria-hidden />}
              title={filtered ? t("No matching runs") : t("No runs")}
              description={filtered ? t("Try another sandbox or status.") : t("Runs appear here when someone presses Run on a sandbox page.")}
            />
          </div>
        ) : (
          <div className="mt-2">
            <Table aria-label={t("Runs")} className="min-w-[900px]">
              <TableHeader>
                <TableColumn className="w-[180px]">{t("Time")}</TableColumn>
                <TableColumn isRowHeader>{t("Sandbox")}</TableColumn>
                <TableColumn className="w-[120px]">{t("Language")}</TableColumn>
                <TableColumn className="w-[130px]">{t("Status")}</TableColumn>
                <TableColumn className="w-[100px]">{t("Duration")}</TableColumn>
                <TableColumn className="w-[170px]">{t("Client")}</TableColumn>
                <TableColumn className="w-[64px]">
                  <span className="sr-only">{t("Actions")}</span>
                </TableColumn>
              </TableHeader>
              <TableBody>
                {loading && data.length === 0
                  ? Array.from({ length: 6 }, (_, index) => (
                      <TableRow key={`skeleton-${index}`} id={`skeleton-${index}`}>
                        <TableCell><Skeleton shape="text" className="h-4 w-32" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-40" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-14" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-16" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-12" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-24" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-8" /></TableCell>
                      </TableRow>
                    ))
                  : data.map((execution) => (
                      <TableRow
                        key={execution.id}
                        id={execution.id}
                        onAction={() => setOpenID(execution.id)}
                        className="cursor-pointer hover:bg-background-secondary-default"
                      >
                        <TableCell>
                          <span className="whitespace-nowrap text-text-secondary tabular-nums">{formatDateTime(execution.createdAt)}</span>
                        </TableCell>
                        <TableCell>
                          <ExecutionSandboxName execution={execution} />
                        </TableCell>
                        <TableCell>
                          <LanguageChip language={execution.language} />
                        </TableCell>
                        <TableCell>
                          <ExecutionStatusChip status={execution.status} />
                        </TableCell>
                        <TableCell>
                          <span className="whitespace-nowrap tabular-nums">{formatDuration(execution.duration)}</span>
                        </TableCell>
                        <TableCell>
                          <span dir="ltr" className="truncate font-mono text-body-2-medium text-text-secondary">{execution.clientAddress}</span>
                        </TableCell>
                        <TableCell>
                          <div className="flex justify-end">
                            <RowAction icon={RiEyeLine} label={t("View")} onClick={() => setOpenID(execution.id)} />
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
            <Pagination page={page} totalPages={totalPages} onChange={(next) => update({ page: next })} />
          </div>
        ) : (
          <div className="h-2" />
        )}
      </section>

      <ExecutionDrawer executionID={openID} onClose={closeDrawer} />
    </>
  );
}
