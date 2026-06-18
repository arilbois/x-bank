import { createFileRoute, Link } from "@tanstack/react-router";
import { useArticles } from "@/hooks/useArticles";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Loading } from "@/components/common/loading";
import { ErrorState } from "@/components/common/error-state";
import { ScoreBadge } from "@/components/common/score-badge";
import { StatusBadge } from "@/components/common/status-badge";
import { Newspaper, TrendingUp, Database, Sparkles } from "lucide-react";

export const Route = createFileRoute("/")({
  component: DashboardPage,
});

function DashboardPage() {
  // Total + per-category counts
  const all = useArticles({ limit: 1 });
  const bytmod = useArticles({ category: "bytmod", limit: 1 });
  const persib = useArticles({ category: "persibWay", limit: 1 });
  const warga = useArticles({ category: "sambatWarga", limit: 1 });
  // Analyzed counts per category (status=analyzed)
  const bytmodAna = useArticles({ category: "bytmod", status: "analyzed", limit: 1 });
  const persibAna = useArticles({ category: "persibWay", status: "analyzed", limit: 1 });
  const wargaAna = useArticles({ category: "sambatWarga", status: "analyzed", limit: 1 });
  const allAna = useArticles({ status: "analyzed", limit: 1 });
  // Recent
  const recent = useArticles({ limit: 8 });

  const stats = [
    {
      label: "Total articles",
      value: all.data?.total ?? "—",
      sub: allAna.data
        ? `${allAna.data.total} analyzed (${Math.round((allAna.data.total / (all.data?.total ?? 1)) * 100)}%)`
        : null,
      icon: Database,
      color: "text-zinc-300",
    },
    {
      label: "bytmod",
      value: bytmod.data?.total ?? "—",
      sub: bytmodAna.data
        ? `${bytmodAna.data.total} analyzed (${Math.round((bytmodAna.data.total / (bytmod.data?.total ?? 1)) * 100)}%)`
        : null,
      icon: Sparkles,
      color: "text-emerald-400",
    },
    {
      label: "persibWay",
      value: persib.data?.total ?? "—",
      sub: persibAna.data
        ? `${persibAna.data.total} analyzed (${Math.round((persibAna.data.total / (persib.data?.total ?? 1)) * 100)}%)`
        : null,
      icon: TrendingUp,
      color: "text-blue-400",
    },
    {
      label: "sambatWarga",
      value: warga.data?.total ?? "—",
      sub: wargaAna.data
        ? `${wargaAna.data.total} analyzed (${Math.round((wargaAna.data.total / (warga.data?.total ?? 1)) * 100)}%)`
        : null,
      icon: Newspaper,
      color: "text-amber-400",
    },
  ];

  const isLoading =
    all.isLoading || bytmod.isLoading || persib.isLoading || warga.isLoading;
  const isError =
    all.isError || bytmod.isError || persib.isError || warga.isError;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <p className="text-sm text-zinc-400">
          Quick overview of scraped content and AI-analyzed output.
        </p>
      </div>

      {isLoading ? (
        <Loading label="Loading dashboard..." />
      ) : isError ? (
        <ErrorState message="Failed to load dashboard data." />
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {stats.map((s) => {
              const Icon = s.icon;
              return (
                <Card key={s.label} className="border-zinc-800 bg-zinc-900">
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-xs font-mono uppercase tracking-wider text-zinc-400">
                      {s.label}
                    </CardTitle>
                    <Icon className={`${s.color} h-4 w-4`} />
                  </CardHeader>
                  <CardContent>
                    <div className="text-3xl font-bold tabular-nums">
                      {s.value}
                    </div>
                    {s.sub && (
                      <p className="mt-1 text-xs text-zinc-500">{s.sub}</p>
                    )}
                  </CardContent>
                </Card>
              );
            })}
          </div>

          <div>
            <div className="mb-3 flex items-center justify-between">
              <h2 className="text-sm font-semibold uppercase tracking-wider text-zinc-400">
                Recent
              </h2>
              <Link
                to="/articles"
                params={{}}
                className="text-xs text-blue-400 hover:underline"
              >
                View all →
              </Link>
            </div>
            {recent.isLoading ? (
              <Loading label="Loading recent..." />
            ) : recent.isError ? (
              <ErrorState message="Failed to load recent articles." />
            ) : (
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                {recent.data?.data.slice(0, 8).map((a) => (
                  <Link
                    key={a.id}
                    to="/articles/$id"
                    params={{ id: a.id }}
                    className="block rounded-md border border-zinc-800 bg-zinc-900 p-3 transition-colors hover:border-zinc-700"
                  >
                    <div className="mb-1 flex items-center gap-2">
                      <ScoreBadge score={a.score} />
                      <StatusBadge status={a.status} />
                      <span className="text-[10px] font-mono text-zinc-500">
                        {a.source_name}
                      </span>
                    </div>
                    <div className="line-clamp-2 text-sm text-zinc-100">
                      {a.title}
                    </div>
                    <div className="mt-1 text-[10px] font-mono text-zinc-500">
                      {a.source_category}
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
