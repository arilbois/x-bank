import { createRootRoute, Outlet, redirect } from "@tanstack/react-router";
import { getToken } from "@/lib/api";
import { AppShell } from "@/components/layout/app-shell";

export const Route = createRootRoute({
  beforeLoad: ({ location }) => {
    const token = getToken();
    const isLogin = location.pathname === "/login";
    if (!token && !isLogin) {
      throw redirect({ to: "/login" });
    }
    if (token && isLogin) {
      throw redirect({ to: "/" });
    }
  },
  component: RootLayout,
});

function RootLayout() {
  const isLogin =
    typeof window !== "undefined" && window.location.pathname === "/login";
  if (isLogin) {
    return <Outlet />;
  }
  return (
    <AppShell>
      <Outlet />
    </AppShell>
  );
}
