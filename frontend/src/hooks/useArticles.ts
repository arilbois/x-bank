import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { api, apiErrorMessage } from "@/lib/api";
import type { ListArticlesResponse, Article, Category } from "@/lib/types";

export interface ArticleFilters {
  category?: Category | "";
  source?: string;
  status?: string;
  search?: string;
  page?: number;
  limit?: number;
}

export function useArticles(filters: ArticleFilters) {
  return useQuery({
    queryKey: ["articles", filters],
    placeholderData: keepPreviousData,
    queryFn: async (): Promise<ListArticlesResponse> => {
      const params: Record<string, string | number> = {
        page: filters.page ?? 1,
        limit: filters.limit ?? 20,
      };
      if (filters.category) params.category = filters.category;
      if (filters.source) params.source = filters.source;
      if (filters.status) params.status = filters.status;
      if (filters.search) params.search = filters.search;
      const { data } = await api.get<ListArticlesResponse>("/articles", {
        params,
      });
      return data;
    },
  });
}

export function useArticleSources(category?: Category | "") {
  return useQuery({
    queryKey: ["article-sources", category ?? ""],
    queryFn: async (): Promise<string[]> => {
      // Backend doesn't expose a /sources endpoint; derive from articles.
      const { data } = await api.get<ListArticlesResponse>("/articles", {
        params: { category: category || undefined, limit: 200 },
      });
      const set = new Set<string>();
      (data.data ?? []).forEach((a: Article) => set.add(a.source_name));
      return Array.from(set).sort();
    },
    staleTime: 60_000,
  });
}

export function articlesErrorMessage(err: unknown): string {
  return apiErrorMessage(err);
}
