import { Badge } from "@/components/ui/badge";
import type { ArticleStatus } from "@/lib/types";

const map: Record<ArticleStatus, { label: string; variant: "secondary" | "info" | "success" | "danger" }> = {
  scraped: { label: "scraped", variant: "secondary" },
  analyzing: { label: "analyzing", variant: "info" },
  analyzed: { label: "analyzed", variant: "success" },
  failed: { label: "failed", variant: "danger" },
};

export function StatusBadge({ status }: { status: string }) {
  const cfg = (map as Record<string, { label: string; variant: "secondary" | "info" | "success" | "danger" }>)[status] ?? {
    label: status,
    variant: "secondary" as const,
  };
  return <Badge variant={cfg.variant}>{cfg.label}</Badge>;
}
