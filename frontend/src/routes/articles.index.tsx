import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useArticles, useArticleSources } from "@/hooks/useArticles";
import { useAuth } from "@/hooks/useAuth";
import { CATEGORIES, type Category } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Loading } from "@/components/common/loading";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { ScoreBadge } from "@/components/common/score-badge";
import { StatusBadge } from "@/components/common/status-badge";
import { formatDate, truncate } from "@/lib/utils";
import { ChevronLeft, ChevronRight, Search, X, Sparkles } from "lucide-react";

export const Route = createFileRoute("/articles/")({
  component: ArticlesPage,
});

const VOICE_LABEL: Record<string, { label: string; emoji: string }> = {
  persibway: { label: "Bobotoh", emoji: "💙" },
  persibWay: { label: "Bobotoh", emoji: "💙" },
  sambatwarga: { label: "Warga Gelisah", emoji: "😤" },
  sambatWarga: { label: "Warga Gelisah", emoji: "😤" },
  bytmod: { label: "InfoSec", emoji: "🔐" },
};

const STATUS_OPTIONS = [
  { value: "", label: "All statuses" },
  { value: "scraped", label: "Scraped (no analysis)" },
  { value: "analyzed", label: "Analyzed" },
];

function voiceFor(category: string) {
  return VOICE_LABEL[category] ?? VOICE_LABEL[category.toLowerCase()] ?? null;
}

function ArticlesPage() {
  const [page, setPage] = useState(1);
  const [category, setCategory] = useState<Category | "">("");
  const [source, setSource] = useState("");
  const [status, setStatus] = useState("");
  const [searchInput, setSearchInput] = useState("");
  const [activeSearch, setActiveSearch] = useState("");
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const qc = useQueryClient();

  const sources = useArticleSources(category || undefined);
  const list = useArticles({
    category,
    source,
    status,
    search: activeSearch,
    page,
    limit: 12,
  });

  const batchAnalyze = useMutation({
    mutationFn: async () => {
      const res = await fetch("/api/analyze/batch", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + (localStorage.getItem("cb_token") ?? ""),
        },
        body: JSON.stringify({ category: category || "", limit: 20 }),
      });
      if (!res.ok) throw new Error("batch analyze failed");
      return res.json();
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["articles"] });
    },
  });

  function onSearchSubmit(e: React.FormEvent) {
    e.preventDefault();
    setActiveSearch(searchInput);
    setPage(1);
  }

  function clearFilters() {
    setCategory("");
    setSource("");
    setStatus("");
    setSearchInput("");
    setActiveSearch("");
    setPage(1);
  }

  const totalPages = list.data?.total
    ? Math.max(1, Math.ceil(list.data.total / 12))
    : 1;

  return (
    <div className="space-y-4">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Articles</h1>
          <p className="text-sm text-zinc-400">
            {list.data
              ? `${list.data.total} total · page ${list.data.page} / ${totalPages}`
              : "Loading..."}
          </p>
        </div>
        {isAdmin && (
          <Button
            size="sm"
            variant="outline"
            className="border-zinc-700"
            onClick={() => batchAnalyze.mutate()}
            disabled={batchAnalyze.isPending}
          >
            <Sparkles className="mr-1.5 h-3.5 w-3.5" />
            {batchAnalyze.isPending
              ? "Analyzing 20..."
              : "Analyze 20 unanalyzed"}
          </Button>
        )}
      </div>

      {/* Filters */}
      <Card className="border-zinc-800 bg-zinc-900">
        <CardContent className="flex flex-wrap items-center gap-2 p-3">
          <div className="flex items-center gap-1">
            <Button
              size="sm"
              variant={category === "" ? "default" : "outline"}
              onClick={() => {
                setCategory(""); setPage(1);
              }}
            >
              All
            </Button>
            {CATEGORIES.map((c) => (
              <Button
                key={c}
                size="sm"
                variant={category === c ? "default" : "outline"}
                onClick={() => {
                  setCategory(c);
                  setPage(1);
                }}
              >
                {c}
              </Button>
            ))}
          </div>

          {sources.data && sources.data.length > 0 && (
            <select
              className="h-8 rounded-md border border-zinc-700 bg-zinc-800 px-2 text-xs text-zinc-100"
              value={source}
              onChange={(e) => {
                setSource(e.target.value);
                setPage(1);
              }}
            >
              <option value="">All sources</option>
              {sources.data.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          )}

          <select
            className="h-8 rounded-md border border-zinc-700 bg-zinc-800 px-2 text-xs text-zinc-100"
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
          >
            {STATUS_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>

          <form onSubmit={onSearchSubmit} className="flex items-center gap-1">
            <Input
              placeholder="Search..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="h-8 w-40 border-zinc-700 bg-zinc-800 text-xs"
            />
            <Button type="submit" size="icon" variant="outline">
              <Search className="h-3.5 w-3.5" />
            </Button>
          </form>

          {(category || source || status || activeSearch) && (
            <Button
              size="sm"
              variant="ghost"
              onClick={clearFilters}
              className="ml-auto"
            >
              <X className="h-3.5 w-3.5" />
              Clear
            </Button>
          )}
        </CardContent>
      </Card>

      {/* List */}
      {list.isLoading ? (
        <Loading label="Loading articles..." />
      ) : list.isError ? (
        <ErrorState message="Failed to load articles." />
      ) : !list.data?.data.length ? (
        <EmptyState
          title="No articles match the current filters."
          description="Try clearing filters or running a scrape from the Scrape page."
        />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          {list.data.data.map((a) => {
            const v = voiceFor(a.source_category);
            return (
              <Link
                key={a.id}
                to="/articles/$id"
                params={{ id: a.id }}
                className="block rounded-md border border-zinc-800 bg-zinc-900 p-4 transition-colors hover:border-zinc-700"
              >
                <div className="mb-2 flex flex-wrap items-center gap-2">
                  <ScoreBadge score={a.score} />
                  <StatusBadge status={a.status} />
                  <span className="rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-300">
                    {a.source_name}
                  </span>
                  {v && (
                    <span
                      className="rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-300"
                      title={`Voice: ${v.label}`}
                    >
                      {v.emoji}
                    </span>
                  )}
                  <span className="ml-auto font-mono text-[10px] text-zinc-500">
                    {formatDate(a.published_at)}
                  </span>
                </div>
                <h3 className="line-clamp-2 text-sm font-medium text-zinc-100">
                  {a.title}
                </h3>
                {a.excerpt && (
                  <p className="mt-1 line-clamp-2 text-xs text-zinc-400">
                    {truncate(a.excerpt, 140)}
                  </p>
                )}
              </Link>
            );
          })}
        </div>
      )}

      {/* Pagination */}
      {list.data && totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 pt-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            Prev
          </Button>
          <span className="font-mono text-xs text-zinc-500">
            {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            Next
            <ChevronRight className="h-3.5 w-3.5" />
          </Button>
        </div>
      )}

      {batchAnalyze.isSuccess && (
        <p className="text-center font-mono text-xs text-emerald-400">
          Batch done: {batchAnalyze.data?.success}/{batchAnalyze.data?.total} analyzed
        </p>
      )}
    </div>
  );
}
