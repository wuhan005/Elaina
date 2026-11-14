import { Navigate, Outlet, useLocation } from "react-router";
import { useIsAuthenticated } from "@/store/auth";
import { AppSidebar } from "./app-sidebar";

export function AppLayout() {
  const isAuthenticated = useIsAuthenticated();
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/sign-in" replace state={{ from: location.pathname }} />;
  }

  return (
    <div className="flex min-h-dvh gap-3 bg-background-full p-3 lg:gap-6 lg:p-4">
      <div className="sticky top-3 h-[calc(100dvh-24px)] shrink-0 lg:top-4 lg:h-[calc(100dvh-32px)]">
        <AppSidebar />
      </div>
      <main className="min-w-0 flex-1 py-3 lg:py-6 lg:pe-4">
        <div className="mx-auto w-full max-w-6xl">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
