import { createFileRoute, Navigate } from "@tanstack/react-router";
import { useState } from "react";
import { useScrapeRun } from "@/hooks/useScrapeRun";
import { useArticles } from "@/hooks/useArticles";
import { CATEGORIES, type Category } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Loading } from "@/components/common/loading";
import { Sparkles, Loader2, CheckCircle2, XCircle } from "lucide-react";
import { formatDate } from "@/lib/utils";
import { useAuth } from "@/hooks/useAuth";

export const Route = createFileRoute("/scrape")({
  component: ScrapePage,
});

function ScrapePage() {
  const { user } = useAuth();
  const [results, setResults] = useState<
    Record<string, { ok: boolean; msg: string } | null>
  >({});
  const scrape = useScrapeRun();

  if (user && user.role !== "admin") {
    return <Navigate to="/" />;
  }

  function run(category: Category | "all") {
    setResults((r) => ({ ...r, [category]: null }));
    scrape.mutate(category as Category, {
      onSuccess: (data) => {
        setResults((r) => ({ ...r, [category]: { ok: true, msg: data.status } }));
      },
      onError: (err) => {
        setResults((r) => ({
          ...r,
          [category]: { ok: false, msg: (err as Error).message },
        }));
      },
    });
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="flex items-center gap-2 text-2xl font-semibold tracking-tight">
          <Sparkles className="h-5 w-5 text-blue-400" />
          Scrape
        </h1>
        <p className="text-sm text-zinc-400">
          Manually trigger scrapers. Cron jobs run automatically on schedule.
        </p>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <ScrapeButton
          label="All sources"
          onClick={() => run("all")}
          loading={scrape.isPending && scrape.variables === "all"}
          result={results.all ?? null}
        />
        {CATEGORIES.map((c) => (
          <ScrapeButton
            key={c}
            label={c}
            onClick={() => run(c)}
            loading={scrape.isPending && scrape.variables === c}
            result={results[c] ?? null}
          />
        ))}
      </div>

      <SourceBreakdown />
    </div>
  );
}

function ScrapeButton({
  label,
  onClick,
  loading,
  result,
}: {
  label: string;
  onClick: () => void;
  loading: boolean;
  result: { ok: boolean; msg: string } | null;
}) {
  return (
    <Card className="border-zinc-800 bg-zinc-900">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm text-zinc-200">{label}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        <Button
          className="w-full"
          onClick={onClick}
          disabled={loading}
        >
          {loading && <Loader2 className="h-4 w-4 animate-spin" />}
          Run scrape
        </Button>
        {result && (
          <div
            className={
              result.ok
                ? "flex items-center gap-1.5 text-xs text-emerald-400"
                : "flex items-center gap-1.5 text-xs text-red-400"
            }
          >
            {result.ok ? (
              <CheckCircle2 className="h-3.5 w-3.5" />
            ) : (
              <XCircle className="h-3.5 w-3.5" />
            )}
            <span className="font-mono">{result.msg}</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function SourceBreakdown() {
  // Use articles list to show per-source counts.
  const all = useArticles({ limit: 1 });
  const bytmod = useArticles({ category: "bytmod", limit: 1 });
  const persib = useArticles({ category: "persibWay", limit: 1 });
  const warga = useArticles({ category: "sambatWarga", limit: 1 });

  if (all.isLoading) return <Loading label="Loading source stats..." />;
  return (
    <Card className="border-zinc-800 bg-zinc-900">
      <CardHeader>
        <CardTitle className="text-sm text-zinc-200">Source totals</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
          <Stat label="All" value={all.data?.total} />
          <Stat label="bytmod" value={bytmod.data?.total} />
          <Stat label="persibWay" value={persib.data?.total} />
          <Stat label="sambatWarga" value={warga.data?.total} />
        </div>
        <p className="mt-3 font-mono text-[10px] text-zinc-500">
          Last query: {formatDate(new Date().toISOString())}
        </p>
      </CardContent>
    </Card>
  );
}

function Stat({ label, value }: { label: string; value?: number }) {
  return (
    <div className="rounded-md border border-zinc-800 bg-zinc-950 p-3">
      <p className="font-mono text-[10px] uppercase tracking-wider text-zinc-500">
        {label}
      </p>
      <p className="text-2xl font-bold tabular-nums">{value ?? "—"}</p>
    </div>
  );
}
