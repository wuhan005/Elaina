import { useCallback } from "react";
import { RouterProvider } from "react-router";
import { Toaster } from "@/components/base/toast/toast";
import { DirectionProvider } from "@/components/foundations/direction/direction";
import { TemplateCopyProvider } from "@/components/foundations/template-copy/template-copy";
import { translate, useLocale } from "@/i18n";
import { router } from "./router";

export function App() {
  const locale = useLocale();
  // A new function per locale makes BoardUI components re-render their built-in copy.
  const translateCopy = useCallback((text: string) => translate(text), [locale]);

  return (
    <DirectionProvider locale={locale}>
      <TemplateCopyProvider translate={translateCopy}>
        <RouterProvider router={router} />
        <Toaster closeLabel={translate("Dismiss notification")} aria-label={translate("Notifications")} />
      </TemplateCopyProvider>
    </DirectionProvider>
  );
}
