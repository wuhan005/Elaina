import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { RiCodeBoxLine, RiExternalLinkLine, RiFileCopyLine, RiPencilLine } from "@remixicon/react";
import type { CreateSandboxReq, Template } from "@/api/api";
import { client } from "@/api/client";
import { PageHeader } from "@/components/application/layout/page-header";
import { FormSection, Panel } from "@/components/application/layout/panel";
import { Button, ButtonLink } from "@/components/base/buttons/button";
import { HintText } from "@/components/base/input/hint-text";
import { Input } from "@/components/base/input/input";
import { Select, SelectItem } from "@/components/base/select/select";
import { Skeleton } from "@/components/base/skeleton/skeleton";
import { SpinnerGlyph } from "@/components/base/spinner/spinner";
import { SwitchCard } from "@/components/base/switch/switch-card";
import { Textarea } from "@/components/base/textarea/textarea";
import { toast } from "@/components/base/toast/toast";
import { LANGUAGES_MAP } from "@/const/template";
import { copyRunnerUrl, runnerUrl } from "./sandboxes";

type FormData = Required<CreateSandboxReq>;
type FormErrors = Partial<Record<keyof FormData, string>>;

const PendingGlyph = ({ className }: { className?: string }) => <SpinnerGlyph size="sm" className={className} />;

function languagesOf(template: Template): string {
  return (template.language ?? []).map((value) => LANGUAGES_MAP[value]?.label ?? value).join(", ");
}

export default function SandboxForm() {
  const navigate = useNavigate();
  const params = useParams();
  const isEdit = params.id !== undefined;
  const id = Number(params.id);

  const [form, setForm] = useState<FormData>({ name: "", templateID: 0, placeholder: "", editable: true });
  const [uid, setUid] = useState("");
  const [templates, setTemplates] = useState<Template[]>([]);
  const [errors, setErrors] = useState<FormErrors>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const requests: Promise<unknown>[] = [
      client.templates.allTemplates().then((res) => setTemplates(res.data.data)),
    ];
    if (isEdit) {
      requests.push(
        client.sandboxes.getSandbox({ id }).then((res) => {
          const sandbox = res.data.data;
          setUid(sandbox.uid);
          setForm({
            name: sandbox.name,
            templateID: sandbox.templateID,
            placeholder: sandbox.placeholder,
            editable: sandbox.editable,
          });
        }),
      );
    }
    Promise.all(requests)
      .catch(() => {
        if (isEdit) navigate("/sandbox", { replace: true });
      })
      .finally(() => setLoading(false));
  }, [id, isEdit, navigate]);

  const update = <K extends keyof FormData>(key: K, value: FormData[K]) => {
    setForm((current) => ({ ...current, [key]: value }));
    if (errors[key]) setErrors((current) => ({ ...current, [key]: undefined }));
  };

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const nextErrors: FormErrors = {};
    if (form.name.trim() === "") nextErrors.name = "Name is required.";
    if (!form.templateID) nextErrors.templateID = "Template is required.";
    setErrors(nextErrors);
    if (Object.values(nextErrors).some(Boolean)) return;

    const payload = { ...form, name: form.name.trim() };
    setSaving(true);
    const request = isEdit
      ? client.sandboxes.updateSandbox({ id }, payload).then((res) => {
          toast.success(res.data.data);
        })
      : client.sandboxes.createSandbox(payload).then((res) => {
          toast.success("Sandbox created", {
            description: runnerUrl(res.data.data.uid),
            actions: [{ label: "Copy link", onClick: () => copyRunnerUrl(res.data.data.uid) }],
          });
          navigate("/sandbox");
        });
    request.catch(() => undefined).finally(() => setSaving(false));
  };

  const title = isEdit ? "Edit sandbox" : "New sandbox";
  const selectedTemplate = templates.find((template) => template.id === form.templateID);

  return (
    <>
      <PageHeader
        title={title}
        description={isEdit ? `Update the runner page of ${form.name || "this sandbox"}.` : "Create a runner page with preset code on top of a template."}
        breadcrumbs={[{ label: "Sandboxes", href: "/sandbox", icon: RiCodeBoxLine }, { label: title }]}
        actions={
          isEdit && uid ? (
            <>
              <Button variant="secondary" leadingIcon={RiFileCopyLine} onClick={() => copyRunnerUrl(uid)}>
                Copy link
              </Button>
              <ButtonLink variant="secondary" leadingIcon={RiExternalLinkLine} href={`/r/${uid}`} target="_blank" rel="noreferrer">
                Open runner
              </ButtonLink>
            </>
          ) : undefined
        }
      />

      <form onSubmit={onSubmit} noValidate>
        <Panel bodyClassName="p-0">
          {loading ? (
            <div className="flex flex-col gap-6 p-6">
              {Array.from({ length: 4 }, (_, index) => (
                <div key={index} className="grid gap-4 md:grid-cols-[260px_1fr] md:gap-10">
                  <div className="flex flex-col gap-2">
                    <Skeleton shape="text" className="h-4 w-28" />
                    <Skeleton shape="text" className="h-3 w-44" />
                  </div>
                  <Skeleton className="h-10 w-full rounded-2lg" />
                </div>
              ))}
            </div>
          ) : (
            <>
              <FormSection title="General" description="The name is only shown in this console.">
                <Input
                  label="Name"
                  placeholder="e.g. Hello world in Go"
                  value={form.name}
                  onChange={(value) => update("name", value)}
                  isRequired
                  isInvalid={errors.name !== undefined}
                  errorMessage={errors.name}
                  autoFocus={!isEdit}
                />
                {isEdit && uid && (
                  <Input label="Runner link" value={runnerUrl(uid)} isReadOnly inputDir="ltr" />
                )}
              </FormSection>

              <FormSection
                title="Template"
                description="Languages, limits and network permission all come from the template."
              >
                <div className="flex flex-col gap-1">
                  <Select
                    label="Template"
                    placeholder="Select a template"
                    isRequired
                    className="w-full"
                    selectedKey={form.templateID || null}
                    onSelectionChange={(key) => update("templateID", Number(key))}
                    isDisabled={templates.length === 0}
                  >
                    {templates.map((template) => (
                      <SelectItem key={template.id} id={template.id} textValue={template.name}>
                        <span className="truncate">{template.name}</span>
                        <span className="truncate text-text-tertiary">{languagesOf(template)}</span>
                      </SelectItem>
                    ))}
                  </Select>
                  {errors.templateID ? (
                    <HintText isInvalid>{errors.templateID}</HintText>
                  ) : templates.length === 0 ? (
                    <p className="pt-px text-caption-1-medium text-text-secondary">
                      No templates yet.{" "}
                      <Link to="/template/new" className="text-text-primary underline underline-offset-2">
                        Create one first
                      </Link>
                      .
                    </p>
                  ) : selectedTemplate ? (
                    <p className="pt-px text-caption-1-medium text-text-secondary">
                      Runs with {languagesOf(selectedTemplate) || "no languages"}.
                    </p>
                  ) : null}
                </div>
              </FormSection>

              <FormSection
                title="Code"
                description="Preset source code shown on the runner page. Its size is capped by the server's source limit."
              >
                <Textarea
                  label="Source code"
                  placeholder={'print("Hello, Elaina!")'}
                  value={form.placeholder}
                  onChange={(value) => update("placeholder", value)}
                  rows={12}
                  inputClassName="font-mono text-body-2-regular"
                  spellCheck="false"
                  dir="ltr"
                />
                <SwitchCard
                  icon={RiPencilLine}
                  title="Editable"
                  description={form.editable ? "Visitors can modify the code before running it" : "Visitors can only run the preset code"}
                  isSelected={form.editable}
                  onChange={(value) => update("editable", value)}
                />
              </FormSection>
            </>
          )}

          <div className="flex flex-col-reverse gap-2 border-t border-separator-border px-5 py-4 sm:flex-row sm:items-center sm:justify-end md:px-6">
            <Button variant="secondary" onClick={() => navigate("/sandbox")}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={loading || saving}
              aria-busy={saving || undefined}
              leadingIcon={saving ? PendingGlyph : undefined}
            >
              {isEdit ? "Save changes" : "Create sandbox"}
            </Button>
          </div>
        </Panel>
      </form>
    </>
  );
}
