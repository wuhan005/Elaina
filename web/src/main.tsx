import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";
import { applyTheme, THEME_STORAGE_KEY } from "@/components/application/theme/theme-toggle";
import { Toaster } from "@/components/base/toast/toast";
import { router } from "./router";
import "./index.css";

// Apply the saved theme before the first paint to avoid a dark mode flash.
applyTheme(window.localStorage.getItem(THEME_STORAGE_KEY) === "dark" ? "dark" : "light", { persist: false });

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <RouterProvider router={router} />
    <Toaster />
  </StrictMode>,
);
