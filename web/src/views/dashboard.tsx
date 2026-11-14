import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router";
import {
  RiAddLine,
  RiArrowRightLine,
  RiCodeBoxLine,
  RiExternalLinkLine,
  RiGlobalLine,
  RiPencilLine,
  RiShareLine,
  RiStackLine,
  RiTranslate2,
} from "@remixicon/react";
import type { Sandbox, Template } from "@/api/api";
import { client } from "@/api/client";
import { StatCards, type Stat } from "@/components/application/dashboard/stat-cards";
import { PageHeader } from "@/components/application/layout/page-header";
import { EmptyState, Panel } from "@/components/application/layout/panel";
import { RowAction } from "@/components/application/layout/row-action";
import { Button } from "@/components/base/buttons/button";
import { LinkButton } from "@/components/base/buttons/link-button";
import { Skeleton } from "@/components/base/skeleton/skeleton";
import { Table, TableBody, TableCell, TableColumn, TableHeader, TableRow } from "@/components/base/table/table";
import { LANGUAGES } from "@/const/template";
import { cx } from "@/utils/cx";
import { formatDate } from "@/utils/format";
import { EditableChip, TemplateName } from "./sandboxes";

// 仪表盘一次性拉取的沙箱数量上限，仅用于统计展示。
const SANDBOX_SAMPLE_SIZE = 500;
const RECENT_COUNT = 5;

const BAR_COLORS = ["bg-chart-1", "bg-chart-2", "bg-chart-3", "bg-chart-4", "bg-chart-5"];

function percent(part: number, whole: number): string {
  return whole === 0 ? "0%" : `${Math.round((part / whole) * 100)}%`;
}

function LanguageCoverage({ templates, loading }: { templates: Template[]; loading: boolean }) {
  const rows = LANGUAGES.map((language) => ({
    ...language,
    count: templates.filter((template) => template.language?.includes(language.value)).length,
  }));
  const max = Math.max(1, ...rows.map((row) => row.count));

  return (
    <Panel title="Language coverage" description="How many templates enable each runtime.">
      <ul className="flex flex-col gap-3.5">
        {rows.map((row, index) => (
          <li key={row.value} className="flex flex-col gap-1.5">
            <div className="flex items-baseline justify-between gap-3">
              <span className="flex min-w-0 items-baseline gap-2">
                <span className="text-body-medium text-text-primary">{row.label}</span>
                <span className="truncate text-caption-1-medium text-text-tertiary">{row.runtime}</span>
              </span>
              {loading ? (
                <Skeleton shape="text" className="h-3.5 w-8" />
              ) : (
                <span className="text-body-medium text-text-secondary tabular-nums">{row.count}</span>
              )}
            </div>
            <div className="h-2 w-full overflow-hidden rounded-full bg-chart-track">
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

function QuickStart() {
  const steps = [
    {
      icon: RiStackLine,
      title: "Create a template",
      description: "Pick languages, a timeout, a memory cap and network permission.",
      to: "/template/new",
    },
    {
      icon: RiCodeBoxLine,
      title: "Create a sandbox",
      description: "Attach preset code to a template and decide if visitors may edit it.",
      to: "/sandbox/new",
    },
    {
      icon: RiShareLine,
      title: "Share the runner",
      description: "Every sandbox gets a /r/:uid page that compiles and runs code in WASM.",
      to: "/sandbox",
    },
  ];

  return (
    <Panel title="Quick start" description="From zero to a shareable code runner." bodyClassName="gap-2 pt-4">
      {steps.map((step, index) => (
        <Link
          key={step.title}
          to={step.to}
          className="group flex items-center gap-3 rounded-2xl p-2.5 outline-none transition-colors hover:bg-background-secondary-default focus-visible:ring-2 focus-visible:ring-border-focus-ring"
        >
          <span className="flex size-10 shrink-0 items-center justify-center rounded-2lg bg-background-tertiary-default text-foreground-icon-primary">
            <step.icon className="size-5" aria-hidden />
          </span>
          <span className="flex min-w-0 flex-1 flex-col">
            <span className="text-body-medium text-text-primary">
              <span className="text-text-tertiary tabular-nums">{index + 1}. </span>
              {step.title}
            </span>
            <span className="text-body-regular text-text-secondary">{step.description}</span>
          </span>
          <RiArrowRightLine
            className="size-5 shrink-0 text-foreground-icon-quaternary transition-transform group-hover:translate-x-0.5 group-hover:text-foreground-icon-secondary rtl:-scale-x-100"
            aria-hidden
          />
        </Link>
      ))}
    </Panel>
  );
}

export default function Dashboard() {
  const navigate = useNavigate();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [sandboxes, setSandboxes] = useState<Sandbox[]>([]);
  const [sandboxTotal, setSandboxTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      client.templates.allTemplates().then((res) => setTemplates(res.data.data)),
      client.sandboxes.listSandboxes({ page: 1, pageSize: SANDBOX_SAMPLE_SIZE }).then((res) => {
        setSandboxes(res.data.data.sandboxes);
        setSandboxTotal(res.data.data.total);
      }),
    ])
      .catch(() => undefined)
      .finally(() => setLoading(false));
  }, []);

  const recent = useMemo(
    () => [...sandboxes].sort((a, b) => b.createdAt.localeCompare(a.createdAt)).slice(0, RECENT_COUNT),
    [sandboxes],
  );

  const networked = templates.filter((template) => template.internetAccess).length;
  const editable = sandboxes.filter((sandbox) => sandbox.editable).length;
  const languagesInUse = new Set(templates.flatMap((template) => template.language ?? [])).size;

  const stats: Stat[] = [
    {
      icon: RiStackLine,
      label: "Templates",
      value: loading ? "–" : templates.length.toLocaleString(),
      delta: `${languagesInUse}/${LANGUAGES.length} runtimes`,
      deltaColor: "neutral",
    },
    {
      icon: RiCodeBoxLine,
      label: "Sandboxes",
      value: loading ? "–" : sandboxTotal.toLocaleString(),
      delta: sandboxTotal > 0 ? "Live" : "None yet",
      deltaColor: sandboxTotal > 0 ? "lime" : "neutral",
    },
    {
      icon: RiPencilLine,
      label: "Editable sandboxes",
      value: loading ? "–" : editable.toLocaleString(),
      delta: percent(editable, sandboxes.length),
      deltaColor: "neutral",
    },
    {
      icon: RiGlobalLine,
      label: "Network-enabled templates",
      value: loading ? "–" : networked.toLocaleString(),
      delta: percent(networked, templates.length),
      deltaColor: networked > 0 ? "rose" : "lime",
    },
  ];

  return (
    <>
      <PageHeader
        title="Dashboard"
        description="An overview of the templates and sandboxes running on this Elaina instance."
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
        <StatCards stats={stats} />

        <div className="grid gap-4 lg:grid-cols-2">
          <LanguageCoverage templates={templates} loading={loading} />
          <QuickStart />
        </div>

        <Panel
          title="Recent sandboxes"
          description="The latest runner pages you created."
          actions={
            <LinkButton href="/sandbox" size="small" trailingIcon={RiArrowRightLine} onClick={(event) => {
              event.preventDefault();
              navigate("/sandbox");
            }}>
              View all
            </LinkButton>
          }
          bodyClassName="px-0 pb-2 pt-4"
        >
          {!loading && recent.length === 0 ? (
            <EmptyState
              icon={<RiTranslate2 className="size-6" aria-hidden />}
              title="Nothing to show yet"
              description="Sandboxes you create will appear here."
              action={
                <Button leadingIcon={RiAddLine} onClick={() => navigate("/sandbox/new")}>
                  New sandbox
                </Button>
              }
            />
          ) : (
            <Table aria-label="Recent sandboxes" className="min-w-[640px]">
              <TableHeader>
                <TableColumn isRowHeader>Name</TableColumn>
                <TableColumn className="w-[200px]">Template</TableColumn>
                <TableColumn className="w-[130px]">Mode</TableColumn>
                <TableColumn className="w-[140px]">Created</TableColumn>
                <TableColumn className="w-[64px]">
                  <span className="sr-only">Actions</span>
                </TableColumn>
              </TableHeader>
              <TableBody>
                {loading
                  ? Array.from({ length: 3 }, (_, index) => (
                      <TableRow key={`skeleton-${index}`} id={`skeleton-${index}`}>
                        <TableCell><Skeleton shape="text" className="h-4 w-40" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-24" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-16" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-20" /></TableCell>
                        <TableCell><Skeleton shape="text" className="h-4 w-8" /></TableCell>
                      </TableRow>
                    ))
                  : recent.map((sandbox) => (
                      <TableRow
                        key={sandbox.id}
                        id={sandbox.id}
                        onAction={() => navigate(`/sandbox/${sandbox.id}`)}
                        className="cursor-pointer hover:bg-background-secondary-default"
                      >
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
                          <span className="whitespace-nowrap text-text-secondary">{formatDate(sandbox.createdAt)}</span>
                        </TableCell>
                        <TableCell>
                          <div className="flex justify-end">
                            <RowAction icon={RiExternalLinkLine} label="Open runner" href={`/r/${sandbox.uid}`} />
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
              </TableBody>
            </Table>
          )}
        </Panel>
      </div>
    </>
  );
}
