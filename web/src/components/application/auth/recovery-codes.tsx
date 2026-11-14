import { RiDownload2Line, RiFileCopyLine } from "@remixicon/react";
import { Button } from "@/components/base/buttons/button";
import { toast } from "@/components/base/toast/toast";
import { useT } from "@/i18n";

export function RecoveryCodes({ codes, username }: { codes: string[]; username: string }) {
  const t = useT();
  const text = codes.join("\n");

  const download = () => {
    const blob = new Blob([`${t("Elaina recovery codes for {username}", { username })}\n\n${text}\n`], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `elaina-recovery-codes-${username}.txt`;
    link.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="flex flex-col gap-3">
      <ul dir="ltr" className="grid grid-cols-2 gap-x-6 gap-y-1.5 rounded-2xl bg-background-secondary-default px-5 py-4 font-mono text-body-medium text-text-primary">
        {codes.map((code) => (
          <li key={code} className="select-all">
            {code}
          </li>
        ))}
      </ul>
      <div className="flex gap-2">
        <Button
          variant="secondary"
          size="small"
          leadingIcon={RiFileCopyLine}
          onClick={() =>
            navigator.clipboard
              .writeText(text)
              .then(() => toast.success(t("Copied")))
              .catch(() => toast.error(t("Could not copy")))
          }
        >
          {t("Copy")}
        </Button>
        <Button variant="secondary" size="small" leadingIcon={RiDownload2Line} onClick={download}>
          {t("Download")}
        </Button>
      </div>
    </div>
  );
}
