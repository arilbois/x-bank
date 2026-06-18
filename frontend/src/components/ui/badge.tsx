import * as React from "react";
import { cn } from "@/lib/utils";

export type BadgeVariant =
  | "default"
  | "secondary"
  | "outline"
  | "success"
  | "warning"
  | "danger"
  | "info";

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
}

const variantClasses: Record<BadgeVariant, string> = {
  default: "bg-blue-600/20 text-blue-300 ring-blue-500/30",
  secondary: "bg-zinc-800 text-zinc-300 ring-zinc-700",
  outline: "bg-transparent text-zinc-300 ring-zinc-700",
  success: "bg-emerald-600/20 text-emerald-300 ring-emerald-500/30",
  warning: "bg-amber-600/20 text-amber-300 ring-amber-500/30",
  danger: "bg-red-600/20 text-red-300 ring-red-500/30",
  info: "bg-sky-600/20 text-sky-300 ring-sky-500/30",
};

export function Badge({
  className,
  variant = "default",
  ...props
}: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-[11px] font-medium ring-1 ring-inset",
        variantClasses[variant],
        className,
      )}
      {...props}
    />
  );
}
