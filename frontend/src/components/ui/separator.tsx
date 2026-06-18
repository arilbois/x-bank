import * as React from "react";
import { cn } from "@/lib/utils";

export const Separator = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & {
    orientation?: "horizontal" | "vertical";
  }
>(function Separator({ className, orientation = "horizontal", ...props }, ref) {
  return (
    <div
      ref={ref}
      role="separator"
      className={cn(
        "bg-zinc-800",
        orientation === "horizontal" ? "h-px w-full" : "h-full w-px",
        className,
      )}
      {...props}
    />
  );
});
