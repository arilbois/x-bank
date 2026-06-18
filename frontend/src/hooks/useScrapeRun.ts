import { useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { Category, ScrapeRunResponse } from "@/lib/types";

export function useScrapeRun() {
  return useMutation({
    mutationFn: async (category: Category | "" | "all") => {
      const { data } = await api.post<ScrapeRunResponse>("/scrape/run", {
        category: category === "" || category === "all" ? "" : category,
      });
      return data;
    },
  });
}
