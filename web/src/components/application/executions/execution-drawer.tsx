import { useEffect, useState, type ReactNode } from "react";
import { Link } from "react-router";
import { RiExternalLinkLine, RiFileCopyLine } from "@remixicon/react";
import type { ExecutionDetail } from "@/api/api";
import { client } from "@/api/client";
import { ButtonLink } from "@/components/base/buttons/button";
import { IconButton } from "@/components/base/buttons/icon-button";
import { Drawer, DrawerBody, DrawerFooter, DrawerHeader } from "@/components/base/drawer/drawer";
import { SegmentedControl, SegmentedControlItem } from "@/components/base/segmented-control/segmented-control";
import { Skeleton } from "@/components/base/skeleton/skeleton";
import { toast } from "@/components/base/toast/toast";
import { ScrollFade, useScrollFade } from "@/components/foundations/scroll-fade/scroll-fade";
import { LANGUAGES_MAP } from "@/const/template";
import { useT } from "@/i18n";
import { formatDateTime } from "@/utils/format";
import { ExecutionStatusChip, formatBytes, formatDuration } from "./execution-status";

type Pane = "code" | "stdout" | "stderr";

function Detail({ label, children, wide }: { label: string; children: ReactNode; wide?: boolean }) {
  return (
    <div className={wide ? "col-span-2 flex min-w-0 flex-col gap-0.5" : "flex min-w-0 flex-col gap-0.5"}>
      <dt className="text-caption-1-medium text-text-tertiary">{label}</dt>
      <dd className="min-w-0 text-body-medium text-text-primary">{children}</dd>
    </div>
  );
}

function TextBlock({ text, empty }: { text: string; empty: string }) {
  const t = useT();
  const fade = useScrollFade();
  if (text === "") {
    return <p className="rounded-2xl bg-background-secondary-default px-4 py-6 text-center text-body-regular text-text-tertiary">{empty}</p>;
  }
  return (
    <div className="relative">
      <pre
        dir="ltr"
        onScroll={fade.onScroll}
        className="max-h-[420px] overflow-auto rounded-2xl bg-background-secondary-default p-4 font-mono text-caption-1-medium whitespace-pre text-text-primary"
      >
        {text}
      </pre>
      <ScrollFade progress={fade.progress} surfaceClassName="from-background-secondary-default" className="rounded-t-2xl" />
      <IconButton
        icon={RiFileCopyLine}
        size="small"
        aria-label={t("Copy")}
        className="absolute top-2 end-2 z-20"
        onClick={() =>
          navigator.clipboard
            .writeText(text)
            .then(() => toast.success(t("Copied")))
            .catch(() => toast.error(t("Could not copy")))
        }
      />
    </div>
  );
}

export function ExecutionDrawer({ executionID, onClose }: { executionID: number | null; onClose: () => void }) {
  const [execution, setExecution] = useState<ExecutionDetail | null>(null);
  const [pane, setPane] = useState<Pane>("code");
  const t = useT();

  useEffect(() => {
    if (executionID === null) return;
    setExecution(null);
    client.executions
      .getExecution({ id: executionID })
      .then((res) => {
        setExecution(res.data.data);
        setPane(res.data.data.code === "" && res.data.data.stdout !== "" ? "stdout" : "code");
      })
      .catch(onClose);
  }, [executionID, onClose]);

  const sandbox = execution?.sandbox;

  return (
    <Drawer isOpen={executionID !== null} onOpenChange={(open) => !open && onClose()} size="lg">
      <DrawerHeader
        title={t("Run #{id}", { id: executionID ?? "" })}
        description={execution ? formatDateTime(execution.createdAt) : undefined}
      />
      <DrawerBody className="flex flex-col gap-5">
        {execution ? (
          <>
            <dl className="grid grid-cols-2 gap-x-6 gap-y-4">
              <Detail label={t("Status")}>
                <ExecutionStatusChip status={execution.status} />
              </Detail>
              <Detail label={t("Sandbox")}>
                {sandbox ? (
                  sandbox.deleted ? (
                    <span className="text-text-secondary">{t("{name} (deleted)", { name: sandbox.name })}</span>
                  ) : (
                    <Link to={`/sandbox/${sandbox.id}`} className="underline-offset-2 hover:underline">
                      {sandbox.name}
                    </Link>
                  )
                ) : (
                  <span className="text-text-tertiary">#{execution.sandboxID}</span>
                )}
              </Detail>
              <Detail label={t("Language")}>{LANGUAGES_MAP[execution.language]?.label ?? (execution.language || "—")}</Detail>
              <Detail label={t("Duration")}>{formatDuration(execution.duration)}</Detail>
              <Detail label={t("Code size")}>{formatBytes(execution.codeBytes)}</Detail>
              <Detail label={t("Client")}>
                <span dir="ltr" className="font-mono text-body-2-medium">{execution.clientAddress || "—"}</span>
              </Detail>
              <Detail label={t("User agent")} wide>
                <span className="break-words text-body-2-regular text-text-secondary">{execution.userAgent || "—"}</span>
              </Detail>
            </dl>

            {execution.error && (
              <div className="flex flex-col gap-1 rounded-2xl bg-notification-error-background px-4 py-3 text-notification-error-foreground">
                <span className="text-caption-1-semibold">{t("Error")}</span>
                <pre dir="ltr" className="font-mono text-caption-1-medium break-words whitespace-pre-wrap">{execution.error}</pre>
              </div>
            )}

            <div className="flex flex-col gap-3">
              <SegmentedControl
                aria-label={t("Content")}
                selectedKeys={new Set([pane])}
                onSelectionChange={(keys) => {
                  const next = [...keys][0];
                  if (next) setPane(next as Pane);
                }}
              >
                <SegmentedControlItem id="code">{t("Code")}</SegmentedControlItem>
                <SegmentedControlItem id="stdout">{t("Output")}</SegmentedControlItem>
                <SegmentedControlItem id="stderr">{t("Stderr")}</SegmentedControlItem>
              </SegmentedControl>

              {pane === "code" && (
                <TextBlock
                  text={execution.code}
                  empty={execution.status === "rejected" ? t("Code isn't stored for rejected requests.") : t("No code.")}
                />
              )}
              {pane === "stdout" && <TextBlock text={execution.stdout} empty={t("No output.")} />}
              {pane === "stderr" && <TextBlock text={execution.stderr} empty={t("Nothing on stderr.")} />}

              {execution.truncated && <p className="text-caption-1-medium text-text-tertiary">{t("Fields longer than 64 KiB were cut.")}</p>}
            </div>
          </>
        ) : (
          <div className="flex flex-col gap-4">
            <Skeleton shape="text" className="h-4 w-2/3" />
            <Skeleton shape="text" className="h-4 w-1/2" />
            <Skeleton className="h-60 w-full rounded-2xl" />
          </div>
        )}
      </DrawerBody>
      {sandbox && !sandbox.deleted && (
        <DrawerFooter>
          <ButtonLink variant="secondary" leadingIcon={RiExternalLinkLine} href={`/r/${sandbox.uid}`} target="_blank" rel="noreferrer">
            {t("Open runner")}
          </ButtonLink>
        </DrawerFooter>
      )}
    </Drawer>
  );
}
