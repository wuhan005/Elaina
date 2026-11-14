import type { ComponentType } from "react";
import { createBrowserRouter, Navigate, useParams } from "react-router";
import { AppLayout } from "@/components/application/layout/app-layout";
import Dashboard from "@/views/dashboard";
import SandboxForm from "@/views/sandbox-form";
import Sandboxes from "@/views/sandboxes";
import SignIn from "@/views/sign-in";
import TemplateForm from "@/views/template-form";
import Templates from "@/views/templates";

// 新建与编辑共用同一表单组件，按 id 设置 key 以便切换时重置表单状态。
function keyedById(Page: ComponentType) {
  return function KeyedPage() {
    const { id } = useParams();
    return <Page key={id ?? "new"} />;
  };
}

const TemplateFormPage = keyedById(TemplateForm);
const SandboxFormPage = keyedById(SandboxForm);

export const router = createBrowserRouter([
  { path: "/sign-in", Component: SignIn },
  {
    Component: AppLayout,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: "dashboard", Component: Dashboard },
      { path: "template", Component: Templates },
      { path: "template/new", Component: TemplateFormPage },
      { path: "template/:id", Component: TemplateFormPage },
      { path: "sandbox", Component: Sandboxes },
      { path: "sandbox/new", Component: SandboxFormPage },
      { path: "sandbox/:id", Component: SandboxFormPage },
      { path: "*", element: <Navigate to="/dashboard" replace /> },
    ],
  },
]);
