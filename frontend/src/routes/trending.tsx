import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useTrending } from "@/hooks/useTrending";
import { CATEGORIES, type Category } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Loading } from "@/components/common/loading";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { ScoreBadge } from "@/components/common/score-badge";
import { TrendingUp } from "lucide-react";
import { truncate } from "@/lib/utils";

export const Route = createFileRoute("/trending")({
  component: TrendingPage,
});

function TrendingPage() {
  const [category, setCategory] = useState<Category | "">("");
  const trending = useTrending(category || undefined, 20);

  return (
    <div className="space-y-4">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-semibold tracking-tight">
            <TrendingUp className="h-5 w-5 text-blue-400" />
            Trending
          </h1>
          <p className="text-sm text-zinc-400">
            Top-scored articles from the last 72 hours.
          </p>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-1">
        <Button
          size="sm"
          variant={category === "" ? "default" : "outline"}
          onClick={() => setCategory("")}
        >
          All
        </Button>
        {CATEGORIES.map((c) => (
          <Button
            key={c}
            size="sm"
            variant={category === c ? "default" : "outline"}
            onClick={() => setCategory(c)}
          >
            {c}
          </Button>
        ))}
      </div>

      {trending.isLoading ? (
        <Loading label="Loading trending..." />
      ) : trending.isError ? (
        <ErrorState message="Failed to load trending articles." />
      ) : !trending.data?.data.length ? (
        <EmptyState
          title="No trending articles yet."
          description="Run a scrape to populate trending content."
        />
      ) : (
        <div className="grid gap-3">
          {trending.data.data.map((a, i) => (
            <Link
              key={a.id}
              to="/articles/$id"
              params={{ id: a.id }}
              className="block"
            >
              <Card className="border-zinc-800 bg-zinc-900 transition-colors hover:border-zinc-700">
                <CardContent className="flex items-start gap-4 p-4">
                  <div className="flex w-8 shrink-0 flex-col items-center text-zinc-500">
                    <span className="font-mono text-lg">#{i + 1}</span>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="mb-1 flex flex-wrap items-center gap-2">
                      <ScoreBadge score={a.score} />
                      <span className="rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-300">
                        {a.source_category}
                      </span>
                      <span className="rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-300">
                        {a.source_name}
                      </span>
                    </div>
                    <h3 className="line-clamp-2 text-sm font-medium text-zinc-100">
                      {a.title}
                    </h3>
                    {a.excerpt && (
                      <p className="mt-1 line-clamp-2 text-xs text-zinc-400">
                        {truncate(a.excerpt, 200)}
                      </p>
                    )}
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
