import { useState } from "react";
import { Link, useRouterState } from "@tanstack/react-router";
import {
  Database,
  LayoutDashboard,
  Newspaper,
  TrendingUp,
  Sparkles,
  LogOut,
  User as UserIcon,
} from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

interface AppShellProps {
  children: React.ReactNode;
}

const NAV: Array<{
  to: string;
  label: string;
  icon: typeof LayoutDashboard;
  adminOnly?: boolean;
}> = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard },
  { to: "/articles", label: "Articles", icon: Newspaper },
  { to: "/trending", label: "Trending", icon: TrendingUp },
  { to: "/scrape", label: "Scrape", icon: Sparkles, adminOnly: true },
];

export function AppShell({ children }: AppShellProps) {
  const { user, logout } = useAuth();
  const location = useRouterState({ select: (s) => s.location.pathname });
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="min-h-screen flex bg-zinc-950 text-zinc-100">
      {/* Sidebar */}
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 w-60 transform border-r border-zinc-800 bg-zinc-900 transition-transform lg:relative lg:translate-x-0",
          sidebarOpen ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <div className="flex h-14 items-center gap-2 border-b border-zinc-800 px-4">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-blue-600/15 text-blue-400">
            <Database className="h-4 w-4" />
          </div>
          <div className="flex flex-col leading-tight">
            <span className="text-sm font-semibold">x-bank</span>
            <span className="text-[10px] text-zinc-500 font-mono">v2 · go</span>
          </div>
        </div>

        <nav className="flex flex-col gap-1 p-3">
          {NAV.map((item) => {
            if (item.adminOnly && user?.role !== "admin") return null;
            const active =
              location === item.to ||
              (item.to !== "/" && location.startsWith(item.to));
            const Icon = item.icon;
            return (
              <Link
                key={item.to}
                to={item.to}
                onClick={() => setSidebarOpen(false)}
                className={cn(
                  "flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors",
                  active
                    ? "bg-blue-600/15 text-blue-300"
                    : "text-zinc-300 hover:bg-zinc-800",
                )}
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="absolute bottom-0 left-0 right-0 border-t border-zinc-800 p-3">
          <div className="mb-2 flex items-center gap-2 px-2 text-xs text-zinc-400">
            <UserIcon className="h-3.5 w-3.5" />
            <span className="truncate">{user?.username ?? "—"}</span>
            <span className="ml-auto rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-300">
              {user?.role ?? "—"}
            </span>
          </div>
          <Button
            variant="outline"
            size="sm"
            className="w-full"
            onClick={logout}
          >
            <LogOut className="h-3.5 w-3.5" />
            Logout
          </Button>
        </div>
      </aside>

      {/* Backdrop for mobile */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/60 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Main */}
      <div className="flex flex-1 flex-col min-w-0">
        <header className="flex h-14 items-center gap-2 border-b border-zinc-800 bg-zinc-900 px-4 lg:px-6">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={() => setSidebarOpen((s) => !s)}
            aria-label="Toggle menu"
          >
            <Database className="h-5 w-5" />
          </Button>
          <div className="text-sm font-mono text-zinc-400">
            x-bank.syahril.site
          </div>
          <div className="ml-auto text-xs text-zinc-500 font-mono">
            {new Date().toISOString().slice(0, 10)}
          </div>
        </header>

        <Separator />

        <main className="flex-1 overflow-auto p-4 lg:p-6">{children}</main>
      </div>
    </div>
  );
}
