import * as React from "react";
import { cn } from "@/lib/utils";

export const ScrollArea = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(function ScrollArea({ className, children, ...props }, ref) {
  return (
    <div
      ref={ref}
      className={cn(
        "relative overflow-y-auto overflow-x-hidden scrollbar-thin scrollbar-thumb-zinc-700 scrollbar-track-transparent",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
});
