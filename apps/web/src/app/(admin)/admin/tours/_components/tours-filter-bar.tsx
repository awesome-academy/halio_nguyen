"use client";

import { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { CategoryListItem } from "@/types/category.types";
import type { TourStatus } from "@/types/tour.types";

interface ToursFilterBarProps {
  categories: CategoryListItem[];
  categoryId: string | undefined;
  status: TourStatus | undefined;
  priceMin: string | undefined;
  priceMax: string | undefined;
  onCategoryChange: (categoryId: string | undefined) => void;
  onStatusChange: (status: string | undefined) => void;
  onPriceChange: (bounds: { price_min?: string; price_max?: string }) => void;
}

const STATUS_OPTIONS: { value: TourStatus; label: string }[] = [
  { value: "draft", label: "Draft" },
  { value: "published", label: "Published" },
  { value: "archived", label: "Archived" },
];

/** FR-202/FR-203: category + status + price-range filters next to the search box. */
export function ToursFilterBar({ categories, categoryId, status, priceMin, priceMax, onCategoryChange, onStatusChange, onPriceChange }: ToursFilterBarProps) {
  const [minDraft, setMinDraft] = useState(priceMin ?? "");
  const [maxDraft, setMaxDraft] = useState(priceMax ?? "");

  useEffect(() => setMinDraft(priceMin ?? ""), [priceMin]);
  useEffect(() => setMaxDraft(priceMax ?? ""), [priceMax]);

  return (
    <>
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

      <Select value={status ?? "all"} onValueChange={(v) => onStatusChange(v === "all" ? undefined : v)}>
        <SelectTrigger className="w-40" aria-label="Status filter">
          <SelectValue placeholder="All statuses" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All statuses</SelectItem>
          {STATUS_OPTIONS.map((o) => (
            <SelectItem key={o.value} value={o.value}>
              {o.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Input
        type="number"
        min={0}
        placeholder="Min price"
        className="w-32"
        aria-label="Minimum price"
        value={minDraft}
        onChange={(e) => setMinDraft(e.target.value)}
        onBlur={() => onPriceChange({ price_min: minDraft || undefined })}
      />
      <Input
        type="number"
        min={0}
        placeholder="Max price"
        className="w-32"
        aria-label="Maximum price"
        value={maxDraft}
        onChange={(e) => setMaxDraft(e.target.value)}
        onBlur={() => onPriceChange({ price_max: maxDraft || undefined })}
      />
    </>
  );
}
