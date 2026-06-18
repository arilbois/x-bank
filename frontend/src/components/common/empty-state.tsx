import type { ReactNode } from "react";
import { Inbox } from "lucide-react";
import { cn } from "@/lib/utils";

export function EmptyState({
  title = "Nothing here",
  description,
  icon,
  action,
  className,
}: {
  title?: string;
  description?: string;
  icon?: ReactNode;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-2 rounded-lg border border-dashed border-zinc-800 bg-zinc-900/30 px-6 py-12 text-center",
        className,
      )}
    >
      <div className="text-zinc-500">
        {icon ?? <Inbox className="h-8 w-8" />}
      </div>
      <p className="text-sm font-medium text-zinc-200">{title}</p>
      {description ? (
        <p className="max-w-sm text-xs text-zinc-500">{description}</p>
      ) : null}
      {action ? <div className="mt-2">{action}</div> : null}
    </div>
  );
}
