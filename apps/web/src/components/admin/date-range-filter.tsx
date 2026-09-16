"use client";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface DateRangeFilterProps {
  /** ISO date (YYYY-MM-DD) or "" for "let the API default it". */
  from: string;
  to: string;
  onChange: (range: { from?: string; to?: string }) => void;
  /** Rendered when the range is empty, to say what the API will default to. */
  emptyHint?: string;
}

/**
 * One inclusive from/to range control, shared by every screen that reports
 * over a period. The values live in the URL (the caller writes them through
 * useListQueryState), so a range survives a reload and can be shared as a
 * link.
 *
 * Native date inputs are deliberate: the range is two plain dates, and a
 * popover calendar would add a dependency and a keyboard-accessibility
 * surface for no gain.
 */
export function DateRangeFilter({ from, to, onChange, emptyHint }: DateRangeFilterProps) {
  return (
    <div className="flex flex-wrap items-end gap-3">
      <div className="space-y-1">
        <Label htmlFor="range-from" className="text-xs text-muted-foreground">
          From
        </Label>
        <Input
          id="range-from"
          type="date"
          value={from}
          max={to || undefined}
          className="h-9 w-[160px]"
          onChange={(e) => onChange({ from: e.target.value })}
        />
      </div>
      <div className="space-y-1">
        <Label htmlFor="range-to" className="text-xs text-muted-foreground">
          To
        </Label>
        <Input
          id="range-to"
          type="date"
          value={to}
          min={from || undefined}
          className="h-9 w-[160px]"
          onChange={(e) => onChange({ to: e.target.value })}
        />
      </div>
      {(from || to) && (
        <button
          type="button"
          className="h-9 text-xs font-medium text-muted-foreground underline-offset-4 hover:underline"
          onClick={() => onChange({ from: undefined, to: undefined })}
        >
          Reset
        </button>
      )}
      {!from && !to && emptyHint && <span className="h-9 text-xs leading-9 text-muted-foreground">{emptyHint}</span>}
    </div>
  );
}
