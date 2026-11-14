import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";
import { applyTheme, THEME_STORAGE_KEY } from "@/components/application/theme/theme-toggle";
import { Toaster } from "@/components/base/toast/toast";
import { router } from "./router";
import "./index.css";

// 首帧前应用已保存的主题，避免深色模式闪烁。
applyTheme(window.localStorage.getItem(THEME_STORAGE_KEY) === "dark" ? "dark" : "light", { persist: false });

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <RouterProvider router={router} />
    <Toaster />
  </StrictMode>,
);
