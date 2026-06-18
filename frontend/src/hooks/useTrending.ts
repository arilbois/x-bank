import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { TrendingResponse, Category } from "@/lib/types";

export function useTrending(category?: Category | "", limit = 20) {
  return useQuery({
    queryKey: ["trending", category ?? "", limit],
    queryFn: async (): Promise<TrendingResponse> => {
      const { data } = await api.get<TrendingResponse>("/trending", {
        params: {
          category: category || undefined,
          limit,
        },
      });
      return data;
    },
  });
}
