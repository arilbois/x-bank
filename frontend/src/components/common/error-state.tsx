import { AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";

export function ErrorState({
  message,
  className,
}: {
  message: string;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex items-start gap-3 rounded-md border border-red-900/50 bg-red-950/40 p-4 text-sm text-red-200",
        className,
      )}
      role="alert"
    >
      <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
      <div>
        <p className="font-medium">Something went wrong</p>
        <p className="text-xs text-red-300/80">{message}</p>
      </div>
    </div>
  );
}
