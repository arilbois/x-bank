import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useArticle, useArticleAnalysis } from "@/hooks/useArticle";
import { useAuth } from "@/hooks/useAuth";
import { api } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Loading } from "@/components/common/loading";
import { ErrorState } from "@/components/common/error-state";
import { ScoreBadge } from "@/components/common/score-badge";
import { StatusBadge } from "@/components/common/status-badge";
import { formatDate } from "@/lib/utils";
import { ArrowLeft, ExternalLink, Sparkles, Copy, Check } from "lucide-react";

export const Route = createFileRoute("/articles/$id")({
  component: ArticleDetailPage,
});

const VOICE_LABEL: Record<string, { label: string; emoji: string }> = {
  persibway: { label: "Bobotoh", emoji: "💙" },
  persibWay: { label: "Bobotoh", emoji: "💙" },
  sambatwarga: { label: "Warga Gelisah", emoji: "😤" },
  sambatWarga: { label: "Warga Gelisah", emoji: "😤" },
  bytmod: { label: "InfoSec", emoji: "🔐" },
};

function voiceFor(category: string) {
  return VOICE_LABEL[category] ?? VOICE_LABEL[category.toLowerCase()] ?? null;
}

function ArticleDetailPage() {
  const { id } = Route.useParams();
  const article = useArticle(id);
  const analysis = useArticleAnalysis(id);
  const { user } = useAuth();
  const qc = useQueryClient();
  const isAdmin = user?.role === "admin";

  const trigger = useMutation({
    mutationFn: async () => {
      const { data } = await api.post<{ data: unknown }>(
        `/articles/${id}/analyze`,
      );
      return data.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["article-analysis", id] });
      qc.invalidateQueries({ queryKey: ["article", id] });
    },
  });

  if (article.isLoading) return <Loading label="Loading article..." />;
  if (article.isError) return <ErrorState message="Failed to load article." />;
  if (!article.data) return <ErrorState message="Article not found." />;

  const a = article.data;
  const voice = voiceFor(a.source_category);

  return (
    <div className="space-y-4">
      <Link
        to="/articles"
        params={{}}
        className="inline-flex items-center gap-1 text-xs text-zinc-400 hover:text-zinc-200"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        Back to articles
      </Link>

      <Card className="border-zinc-800 bg-zinc-900">
        <CardHeader>
          <div className="mb-2 flex flex-wrap items-center gap-2">
            <ScoreBadge score={a.score} />
            <StatusBadge status={a.status} />
            <span className="rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-300">
              {a.source_category}
            </span>
            <span className="rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-300">
              {a.source_name}
            </span>
            {voice && (
              <span className="rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-300">
                voice: {voice.emoji} {voice.label}
              </span>
            )}
            <span className="ml-auto font-mono text-[10px] text-zinc-500">
              {formatDate(a.published_at)}
            </span>
          </div>
          <CardTitle className="text-xl text-zinc-100">{a.title}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {a.excerpt && (
            <p className="text-sm leading-relaxed text-zinc-300">{a.excerpt}</p>
          )}
          {a.url && (
            <a
              href={a.url}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-xs text-blue-400 hover:underline"
            >
              <ExternalLink className="h-3.5 w-3.5" />
              Open original
            </a>
          )}
          {a.tags && a.tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {a.tags.map((t) => (
                <span
                  key={t}
                  className="rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-400"
                >
                  #{t}
                </span>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-zinc-800 bg-zinc-900">
        <CardHeader>
          <div className="flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-blue-400" />
            <CardTitle className="text-sm text-zinc-200">AI Analysis</CardTitle>
            {analysis.data && (
              <span className="rounded bg-emerald-900/30 px-1.5 py-0.5 font-mono text-[10px] text-emerald-300">
                ready
              </span>
            )}
            {!analysis.data && !analysis.isLoading && (
              <span className="rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-500">
                not yet analyzed
              </span>
            )}
            {isAdmin && (
              <Button
                size="sm"
                variant="outline"
                className="ml-auto border-zinc-700 text-xs"
                onClick={() => trigger.mutate()}
                disabled={trigger.isPending}
              >
                {trigger.isPending
                  ? "Analyzing..."
                  : analysis.data
                    ? "Re-analyze"
                    : "Analyze now"}
              </Button>
            )}
          </div>
          {trigger.isError && (
            <p className="mt-2 text-xs text-red-400">
              {(trigger.error as { response?: { data?: { error?: string } } })
                ?.response?.data?.error ?? "Analyze failed."}
            </p>
          )}
        </CardHeader>
        <CardContent>
          {analysis.isLoading ? (
            <Loading label="Loading analysis..." />
          ) : !analysis.data ? (
            <p className="text-xs text-zinc-500">
              No analysis has been generated for this article yet.
              {isAdmin ? ' Click "Analyze now" to run the AI pass.' : ""}
            </p>
          ) : (
            <div className="space-y-3">
              <Field label="Sentiment" value={analysis.data.sentiment} />
              <Field
                label="Summary"
                value={analysis.data.summary}
                multiline
              />
              <CopyableField label="Hook" value={analysis.data.hook} />
              <CopyableField label="Tweet" value={analysis.data.tweet} />
              <CopyableField
                label="Thread opener"
                value={analysis.data.thread_opener}
                multiline
              />
              <p className="font-mono text-[10px] text-zinc-500">
                model: {analysis.data.model}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function Field({
  label,
  value,
  multiline,
}: {
  label: string;
  value: string;
  multiline?: boolean;
}) {
  return (
    <div>
      <p className="mb-1 font-mono text-[10px] uppercase tracking-wider text-zinc-500">
        {label}
      </p>
      <p
        className={
          multiline
            ? "whitespace-pre-wrap text-sm text-zinc-200"
            : "text-sm text-zinc-200"
        }
      >
        {value || "—"}
      </p>
    </div>
  );
}

function CopyableField({
  label,
  value,
  multiline,
}: {
  label: string;
  value: string;
  multiline?: boolean;
}) {
  const [copied, setCopied] = useState(false);
  function onCopy() {
    if (!value) return;
    navigator.clipboard
      .writeText(value)
      .then(() => {
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
      })
      .catch(() => {});
  }
  return (
    <div>
      <div className="mb-1 flex items-center gap-2">
        <p className="font-mono text-[10px] uppercase tracking-wider text-zinc-500">
          {label}
        </p>
        <button
          type="button"
          onClick={onCopy}
          className="inline-flex items-center gap-1 rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[10px] text-zinc-400 hover:bg-zinc-700 hover:text-zinc-200"
          aria-label={`Copy ${label}`}
        >
          {copied ? (
            <>
              <Check className="h-3 w-3" /> copied
            </>
          ) : (
            <>
              <Copy className="h-3 w-3" /> copy
            </>
          )}
        </button>
      </div>
      <p
        className={
          multiline
            ? "whitespace-pre-wrap text-sm text-zinc-200"
            : "text-sm text-zinc-200"
        }
      >
        {value || "—"}
      </p>
    </div>
  );
}
