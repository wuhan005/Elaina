import type { ComponentType } from "react";
import { createBrowserRouter, Navigate, useParams } from "react-router";
import { AppLayout } from "@/components/application/layout/app-layout";
import Account from "@/views/account";
import Dashboard from "@/views/dashboard";
import Executions from "@/views/executions";
import SandboxForm from "@/views/sandbox-form";
import Sandboxes from "@/views/sandboxes";
import SignIn from "@/views/sign-in";
import TemplateForm from "@/views/template-form";
import Templates from "@/views/templates";
import Users from "@/views/users";

// Create and edit share one form component; keying by id resets its state when switching.
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
      { path: "execution", Component: Executions },
      { path: "users", Component: Users },
      { path: "account", Component: Account },
      { path: "*", element: <Navigate to="/dashboard" replace /> },
    ],
  },
]);
