import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import dayjs from "dayjs";
import { Focusable } from "react-aria-components";
import { RiAddLine, RiArrowRightLine, RiCodeBoxLine, RiPulseLine, RiStackLine, RiTimeLine } from "@remixicon/react";
import type { ExecutionStats, ExecutionSummary, Sandbox, Template } from "@/api/api";
import { client } from "@/api/client";
import { StatCards, type Stat } from "@/components/application/dashboard/stat-cards";
import { ExecutionDrawer } from "@/components/application/executions/execution-drawer";
import { ExecutionStatusChip, formatDuration } from "@/components/application/executions/execution-status";
import { PageHeader } from "@/components/application/layout/page-header";
import { EmptyState, Panel } from "@/components/application/layout/panel";
import { Button } from "@/components/base/buttons/button";
import { LinkButton } from "@/components/base/buttons/link-button";
import { Skeleton } from "@/components/base/skeleton/skeleton";
import { Table, TableBody, TableCell, TableColumn, TableHeader, TableRow } from "@/components/base/table/table";
import { Tooltip, TooltipTrigger } from "@/components/base/tooltip/tooltip";
import { LANGUAGES } from "@/const/template";
import { cx } from "@/utils/cx";
import { ExecutionSandboxName, LanguageChip } from "./executions";

// Maximum number of sandboxes fetched for the dashboard statistics.
const SANDBOX_SAMPLE_SIZE = 500;
const RECENT_COUNT = 6;
const CHART_DAYS = 14;

const BAR_COLORS = ["bg-chart-1", "bg-chart-2", "bg-chart-3", "bg-chart-4", "bg-chart-5"];

function createdThisWeek(items: { createdAt: string }[]): number {
  const since = dayjs().subtract(7, "day");
  return items.filter((item) => dayjs(item.createdAt).isAfter(since)).length;
}

function weekDelta(count: number): Pick<Stat, "delta" | "deltaColor"> {
  return count > 0 ? { delta: `+${count} this week`, deltaColor: "lime" } : { delta: "None this week", deltaColor: "neutral" };
}

function RunsChart({ stats, loading }: { stats: ExecutionStats | null; loading: boolean }) {
  const daily = stats?.daily ?? [];
  const max = Math.max(1, ...daily.map((day) => day.total));

  return (
    <Panel
      title="Runs"
      description={`Last ${CHART_DAYS} days`}
      actions={
        <dl className="flex items-center gap-4 text-body-2-medium text-text-secondary">
          <div className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-chart-6" aria-hidden />
            <dt>Succeeded</dt>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-chart-3" aria-hidden />
            <dt>Other</dt>
          </div>
        </dl>
      }
    >
      {loading ? (
        <Skeleton className="h-44 w-full rounded-2xl" />
      ) : (
        <div className="flex flex-col gap-2">
          <div className="flex h-40 items-end gap-1.5 sm:gap-2" role="list" aria-label="Runs per day">
            {daily.map((day) => {
              const succeeded = day.total - day.failed;
              const date = dayjs(day.date);
              const label = `${date.format("MMM D")}: ${day.total} ${day.total === 1 ? "run" : "runs"}, ${day.failed} not successful`;
              return (
                <TooltipTrigger key={day.date} delay={100}>
                  <Focusable>
                    <div
                      role="listitem"
                      tabIndex={0}
                      aria-label={label}
                      className="group flex h-full min-w-0 flex-1 cursor-default flex-col justify-end rounded-md outline-none focus-visible:ring-2 focus-visible:ring-border-focus-ring"
                    >
                      {day.total === 0 ? (
                        <span className="h-1 w-full rounded-full bg-chart-track" />
                      ) : (
                        <span
                          className="flex w-full flex-col overflow-hidden rounded-md transition-[height] duration-500 ease-out"
                          style={{ height: `${Math.max(4, (day.total / max) * 100)}%` }}
                        >
                          {day.failed > 0 && <span className="bg-chart-3 group-hover:bg-chart-3-active" style={{ flexGrow: day.failed }} />}
                          {succeeded > 0 && <span className="bg-chart-6 group-hover:bg-chart-6-active" style={{ flexGrow: succeeded }} />}
                        </span>
                      )}
                    </div>
                  </Focusable>
                  <Tooltip size="md">
                    <span className="flex flex-col">
                      <span>{date.format("ddd, MMM D")}</span>
                      <span className="text-text-secondary tabular-nums">
                        {day.total} total · {succeeded} succeeded · {day.failed} other
                      </span>
                    </span>
                  </Tooltip>
                </TooltipTrigger>
              );
            })}
          </div>
          <div className="flex gap-1.5 sm:gap-2" aria-hidden>
            {daily.map((day, index) => (
              <span key={day.date} className="min-w-0 flex-1 text-center text-caption-2-medium text-text-tertiary tabular-nums">
                {index % 2 === daily.length % 2 ? "" : dayjs(day.date).format("D")}
              </span>
            ))}
          </div>
        </div>
      )}
    </Panel>
  );
}

function Languages({ templates, loading }: { templates: Template[]; loading: boolean }) {
  const rows = LANGUAGES.map((language) => ({
    ...language,
    count: templates.filter((template) => template.language?.includes(language.value)).length,
  }));
  const max = Math.max(1, ...rows.map((row) => row.count));

  return (
    <Panel title="Templates by language">
      <ul className="flex flex-col gap-3.5">
        {rows.map((row, index) => (
          <li key={row.value} className="flex flex-col gap-1.5">
            <div className="flex items-baseline justify-between gap-3">
              <span className="flex min-w-0 items-baseline gap-2">
                <span className="text-body-medium text-text-primary">{row.label}</span>
                <span className="truncate text-caption-1-medium text-text-tertiary">{row.runtime}</span>
              </span>
              {loading ? (
                <Skeleton shape="text" className="h-3.5 w-6" />
              ) : (
                <span className="text-body-medium text-text-secondary tabular-nums">{row.count}</span>
              )}
            </div>
            <div className="h-1.5 w-full overflow-hidden rounded-full bg-chart-track">
              <div
                className={cx("h-full rounded-full transition-[width] duration-700 ease-out", BAR_COLORS[index % BAR_COLORS.length])}
                style={{ width: loading ? "0%" : `${(row.count / max) * 100}%` }}
              />
            </div>
          </li>
        ))}
      </ul>
    </Panel>
  );
}

export default function Dashboard() {
  const navigate = useNavigate();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [sandboxes, setSandboxes] = useState<Sandbox[]>([]);
  const [sandboxTotal, setSandboxTotal] = useState(0);
  const [stats, setStats] = useState<ExecutionStats | null>(null);
  const [recent, setRecent] = useState<ExecutionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [openID, setOpenID] = useState<number | null>(null);

  useEffect(() => {
    Promise.all([
      client.templates.allTemplates().then((res) => setTemplates(res.data.data)),
      client.sandboxes.listSandboxes({ page: 1, pageSize: SANDBOX_SAMPLE_SIZE }).then((res) => {
        setSandboxes(res.data.data.sandboxes);
        setSandboxTotal(res.data.data.total);
      }),
      client.executions
        .executionStats({ days: CHART_DAYS, timezone: Intl.DateTimeFormat().resolvedOptions().timeZone })
        .then((res) => setStats(res.data.data)),
      client.executions.listExecutions({ page: 1, pageSize: RECENT_COUNT }).then((res) => setRecent(res.data.data.executions)),
    ])
      .catch(() => undefined)
      .finally(() => setLoading(false));
  }, []);

  const value = (count: number) => (loading ? "–" : count.toLocaleString());
  const runsThisWeek = (stats?.daily ?? []).slice(-7).reduce((sum, day) => sum + day.total, 0);
  const failed24h = stats?.failed24h ?? 0;

  const statCards: Stat[] = [
    { icon: RiStackLine, label: "Templates", value: value(templates.length), ...weekDelta(createdThisWeek(templates)) },
    { icon: RiCodeBoxLine, label: "Sandboxes", value: value(sandboxTotal), ...weekDelta(createdThisWeek(sandboxes)) },
    {
      icon: RiTimeLine,
      label: "Runs in 24 hours",
      value: value(stats?.last24h ?? 0),
      delta: failed24h > 0 ? `${failed24h} not successful` : "All succeeded",
      deltaColor: failed24h > 0 ? "rose" : "neutral",
    },
    { icon: RiPulseLine, label: "Total runs", value: value(stats?.total ?? 0), ...weekDelta(runsThisWeek) },
  ];

  return (
    <>
      <PageHeader
        title="Dashboard"
        actions={
          <>
            <Button variant="secondary" leadingIcon={RiStackLine} onClick={() => navigate("/template/new")}>
              New template
            </Button>
            <Button leadingIcon={RiAddLine} onClick={() => navigate("/sandbox/new")}>
              New sandbox
            </Button>
          </>
        }
      />

      <div className="flex flex-col gap-4">
        <StatCards stats={statCards} />

        <RunsChart stats={stats} loading={loading} />

        <div className="grid gap-4 lg:grid-cols-3">
          <Panel
            title="Recent runs"
            className="lg:col-span-2"
            actions={
              <LinkButton
                href="/execution"
                size="small"
                trailingIcon={RiArrowRightLine}
                onClick={(event) => {
                  event.preventDefault();
                  navigate("/execution");
                }}
              >
                All runs
              </LinkButton>
            }
            bodyClassName="px-0 pb-2 pt-4"
          >
            {!loading && recent.length === 0 ? (
              <EmptyState icon={<RiPulseLine className="size-6" aria-hidden />} title="No runs" description="Runs appear here when someone presses Run on a sandbox page." />
            ) : (
              <Table aria-label="Recent runs" className="min-w-[560px]">
                <TableHeader>
                  <TableColumn isRowHeader>Sandbox</TableColumn>
                  <TableColumn className="w-[110px]">Language</TableColumn>
                  <TableColumn className="w-[120px]">Status</TableColumn>
                  <TableColumn className="w-[90px]">Duration</TableColumn>
                  <TableColumn className="w-[110px]">When</TableColumn>
                </TableHeader>
                <TableBody>
                  {loading
                    ? Array.from({ length: 3 }, (_, index) => (
                        <TableRow key={`skeleton-${index}`} id={`skeleton-${index}`}>
                          <TableCell><Skeleton shape="text" className="h-4 w-40" /></TableCell>
                          <TableCell><Skeleton shape="text" className="h-4 w-14" /></TableCell>
                          <TableCell><Skeleton shape="text" className="h-4 w-16" /></TableCell>
                          <TableCell><Skeleton shape="text" className="h-4 w-12" /></TableCell>
                          <TableCell><Skeleton shape="text" className="h-4 w-16" /></TableCell>
                        </TableRow>
                      ))
                    : recent.map((execution) => (
                        <TableRow
                          key={execution.id}
                          id={execution.id}
                          onAction={() => setOpenID(execution.id)}
                          className="cursor-pointer hover:bg-background-secondary-default"
                        >
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
                            <span className="whitespace-nowrap text-text-secondary" title={execution.createdAt}>
                              {dayjs(execution.createdAt).format("MMM D, HH:mm")}
                            </span>
                          </TableCell>
                        </TableRow>
                      ))}
                </TableBody>
              </Table>
            )}
          </Panel>

          <Languages templates={templates} loading={loading} />
        </div>
      </div>

      <ExecutionDrawer executionID={openID} onClose={() => setOpenID(null)} />
    </>
  );
}
