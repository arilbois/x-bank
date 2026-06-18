import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { Article, ArticleAnalysis, SingleArticleResponse } from "@/lib/types";

export function useArticle(id: string | undefined) {
  return useQuery({
    queryKey: ["article", id],
    enabled: !!id,
    queryFn: async (): Promise<Article> => {
      const { data } = await api.get<SingleArticleResponse>(`/articles/${id}`);
      return data.data;
    },
  });
}

export function useArticleAnalysis(articleId: string | undefined) {
  return useQuery({
    queryKey: ["article-analysis", articleId],
    enabled: !!articleId,
    // Backend exposes BOTH /analysis/:id and /articles/:id/analysis.
    // /articles/:id/analysis takes the article id directly, so use that.
    queryFn: async (): Promise<ArticleAnalysis | null> => {
      try {
        const { data } = await api.get<{ data: ArticleAnalysis }>(
          `/articles/${articleId}/analysis`,
        );
        return data.data;
      } catch (err: unknown) {
        // 404 = no analysis yet, not an error worth surfacing.
        const status = (err as { response?: { status?: number } })?.response
          ?.status;
        if (status === 404) return null;
        throw err;
      }
    },
  });
}
