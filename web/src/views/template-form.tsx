import { useEffect, useState, type FormEvent } from "react";
import { useNavigate, useParams } from "react-router";
import { RiGlobalLine, RiInformationLine, RiStackLine } from "@remixicon/react";
import type { CreateTemplateReq } from "@/api/api";
import { client } from "@/api/client";
import { PageHeader } from "@/components/application/layout/page-header";
import { FormSection, Panel } from "@/components/application/layout/panel";
import { Button } from "@/components/base/buttons/button";
import { CheckboxCard } from "@/components/base/checkbox/checkbox-card";
import { HintText } from "@/components/base/input/hint-text";
import { Input } from "@/components/base/input/input";
import { NumberField } from "@/components/base/number-field/number-field";
import { Skeleton } from "@/components/base/skeleton/skeleton";
import { SpinnerGlyph } from "@/components/base/spinner/spinner";
import { SwitchCard } from "@/components/base/switch/switch-card";
import { toast } from "@/components/base/toast/toast";
import { LANGUAGES } from "@/const/template";

type FormData = Required<CreateTemplateReq>;
type FormErrors = Partial<Record<keyof FormData, string>>;

const PendingGlyph = ({ className }: { className?: string }) => <SpinnerGlyph size="sm" className={className} />;

function validate(form: FormData): FormErrors {
  const errors: FormErrors = {};
  if (form.name.trim() === "") errors.name = "Name is required.";
  if (form.language.length === 0) errors.language = "Select at least one programming language.";
  if (!Number.isInteger(form.timeout) || form.timeout < 0 || form.timeout > 60) {
    errors.timeout = "Use a whole number from 0 to 60.";
  }
  if (!Number.isInteger(form.maxMemory) || form.maxMemory < 0 || form.maxMemory > 4096) {
    errors.maxMemory = "Use a whole number from 0 to 4096.";
  }
  return errors;
}

export default function TemplateForm() {
  const navigate = useNavigate();
  const params = useParams();
  const isEdit = params.id !== undefined;
  const id = Number(params.id);

  const [form, setForm] = useState<FormData>({
    name: "",
    language: [],
    timeout: 60,
    maxMemory: 512,
    internetAccess: false,
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [loading, setLoading] = useState(isEdit);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!isEdit) return;
    client.templates
      .getTemplate({ id })
      .then((res) => {
        const template = res.data.data;
        setForm({
          name: template.name,
          language: template.language ?? [],
          timeout: template.timeout,
          maxMemory: template.maxMemory,
          internetAccess: template.internetAccess,
        });
      })
      .catch(() => navigate("/template", { replace: true }))
      .finally(() => setLoading(false));
  }, [id, isEdit, navigate]);

  const update = <K extends keyof FormData>(key: K, value: FormData[K]) => {
    setForm((current) => ({ ...current, [key]: value }));
    if (errors[key]) setErrors((current) => ({ ...current, [key]: undefined }));
  };

  const toggleLanguage = (value: string, selected: boolean) => {
    update(
      "language",
      selected ? [...form.language, value] : form.language.filter((item) => item !== value),
    );
  };

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const nextErrors = validate(form);
    setErrors(nextErrors);
    if (Object.values(nextErrors).some(Boolean)) return;

    const payload = { ...form, name: form.name.trim() };
    setSaving(true);
    const request = isEdit
      ? client.templates.updateTemplate({ id }, payload).then((res) => {
          toast.success(res.data.data);
        })
      : client.templates.createTemplate(payload).then(() => {
          toast.success("Template created");
          navigate("/template");
        });
    request.catch(() => undefined).finally(() => setSaving(false));
  };

  const title = isEdit ? "Edit template" : "New template";

  return (
    <>
      <PageHeader
        title={title}
        breadcrumbs={[{ label: "Templates", href: "/template", icon: RiStackLine }, { label: title }]}
      />

      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        <div className="flex items-start gap-3 rounded-2xl bg-background-secondary-default px-4 py-3 text-text-secondary">
          <RiInformationLine className="mt-0.5 size-5 shrink-0 text-foreground-icon-secondary" aria-hidden />
          <p className="text-body-regular">
            Server configuration caps memory, global concurrency, input and workspace usage.
          </p>
        </div>

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
              <FormSection title="General">
                <Input
                  label="Name"
                  placeholder="e.g. Python playground"
                  value={form.name}
                  onChange={(value) => update("name", value)}
                  isRequired
                  isInvalid={errors.name !== undefined}
                  errorMessage={errors.name}
                  autoFocus={!isEdit}
                />
              </FormSection>

              <FormSection
                title="Languages"
                description="Code runs as WebAssembly. Native extensions and subprocesses are unavailable."
              >
                <div className="flex flex-col gap-1">
                  <div className="grid gap-2 sm:grid-cols-2" role="group" aria-label="Programming languages">
                    {LANGUAGES.map((language) => (
                      <CheckboxCard
                        key={language.value}
                        title={language.label}
                        description={language.runtime}
                        isSelected={form.language.includes(language.value)}
                        onChange={(selected) => toggleLanguage(language.value, selected)}
                      />
                    ))}
                  </div>
                  {errors.language ? <HintText isInvalid>{errors.language}</HintText> : null}
                </div>
              </FormSection>

              <FormSection title="Limits">
                <div className="grid gap-5 lg:grid-cols-2">
                  <NumberField
                    label="Execution timeout (s)"
                    value={form.timeout}
                    onChange={(value) => update("timeout", value)}
                    minValue={0}
                    maxValue={60}
                    step={1}
                    formatOptions={{ maximumFractionDigits: 0 }}
                    isRequired
                    isInvalid={errors.timeout !== undefined}
                    hint={errors.timeout ?? "0 uses the 10-second default. Includes compilation and startup; 60s is recommended for C/Go cold starts."}
                  />
                  <NumberField
                    label="WASM linear memory (MiB)"
                    value={form.maxMemory}
                    onChange={(value) => update("maxMemory", value)}
                    minValue={0}
                    maxValue={4096}
                    step={64}
                    formatOptions={{ maximumFractionDigits: 0, useGrouping: false }}
                    isRequired
                    isInvalid={errors.maxMemory !== undefined}
                    hint={errors.maxMemory ?? "0 inherits the server ceiling; nonzero values are capped by it. Start with 512 MiB for C/Go, 128 MiB for interpreters."}
                  />
                </div>
              </FormSection>

              <FormSection
                title="Network"
                description="Compilation is always offline."
              >
                <SwitchCard
                  icon={RiGlobalLine}
                  title="Network access"
                  description={form.internetAccess ? "Supported networking is allowed" : "Sockets and DNS are blocked"}
                  isSelected={form.internetAccess}
                  onChange={(value) => update("internetAccess", value)}
                />
                <p className="-mt-2 text-caption-1-medium text-text-secondary">
                  Supported: Go HTTP/DNS/HTTPS, C IPv4 TCP/DNS without bundled TLS, PHP IPv4 TCP/HTTP. Python and JavaScript have no networking.
                </p>
              </FormSection>
            </>
          )}

          <div className="flex flex-col-reverse gap-2 border-t border-separator-border px-5 py-4 sm:flex-row sm:items-center sm:justify-end md:px-6">
            <Button variant="secondary" onClick={() => navigate("/template")}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={loading || saving}
              aria-busy={saving || undefined}
              leadingIcon={saving ? PendingGlyph : undefined}
            >
              {isEdit ? "Save changes" : "Create template"}
            </Button>
          </div>
        </Panel>
      </form>
    </>
  );
}
