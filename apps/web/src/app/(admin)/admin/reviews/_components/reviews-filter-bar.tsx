"use client";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { ReviewCategory, ReviewStatus } from "@/types/review.types";

interface ReviewsFilterBarProps {
  status: ReviewStatus | undefined;
  categoryId: string | undefined;
  /** A7's seeded lookup — never Phase 3's tour-category endpoint. */
  categories: ReviewCategory[];
  onStatusChange: (status: string | undefined) => void;
  onCategoryChange: (categoryId: string | undefined) => void;
}

/** FR-001: status + review-category filters, combinable with the title search. */
export function ReviewsFilterBar({ status, categoryId, categories, onStatusChange, onCategoryChange }: ReviewsFilterBarProps) {
  return (
    <>
      <Select value={status ?? "all"} onValueChange={(v) => onStatusChange(v === "all" ? undefined : v)}>
        <SelectTrigger className="w-40" aria-label="Status filter">
          <SelectValue placeholder="All statuses" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All statuses</SelectItem>
          <SelectItem value="draft">Draft</SelectItem>
          <SelectItem value="published">Published</SelectItem>
          <SelectItem value="hidden">Hidden</SelectItem>
        </SelectContent>
      </Select>

      <Select value={categoryId ?? "all"} onValueChange={(v) => onCategoryChange(v === "all" ? undefined : v)}>
        <SelectTrigger className="w-48" aria-label="Category filter">
          <SelectValue placeholder="All categories" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All categories</SelectItem>
          {categories.map((c) => (
            <SelectItem key={c.id} value={c.id}>
              {c.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </>
  );
}
