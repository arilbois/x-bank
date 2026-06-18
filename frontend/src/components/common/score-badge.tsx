import { Badge } from "@/components/ui/badge";

export function ScoreBadge({ score }: { score: number }) {
  const s = Math.round(score);
  let variant: "success" | "warning" | "danger" | "secondary" = "secondary";
  if (s >= 80) variant = "success";
  else if (s >= 50) variant = "warning";
  else if (s > 0) variant = "danger";

  return (
    <Badge variant={variant} className="font-mono tabular-nums">
      {s}
    </Badge>
  );
}
