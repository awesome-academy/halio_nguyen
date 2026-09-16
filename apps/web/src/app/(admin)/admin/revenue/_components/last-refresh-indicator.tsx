"use client";

import { Clock } from "lucide-react";

interface LastRefreshIndicatorProps {
  /** null whenever the API process has not completed a refresh (DEC-001). */
  lastRefreshedAt: string | null | undefined;
}

/**
 * DEC-001 / L3: nothing persists the last-refresh time. It lives in the API
 * process's memory, so it is null after a restart and two replicas will
 * disagree. "Unknown" is the correct thing to show then — a fabricated or
 * stale time would be worse than no time at all.
 */
export function LastRefreshIndicator({ lastRefreshedAt }: LastRefreshIndicatorProps) {
  const label = lastRefreshedAt
    ? new Intl.DateTimeFormat("en-US", { dateStyle: "medium", timeStyle: "short" }).format(new Date(lastRefreshedAt))
    : "unknown";

  return (
    <span className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
      <Clock className="h-3.5 w-3.5" />
      Last refresh: <span className="font-medium text-foreground">{label}</span>
    </span>
  );
}
