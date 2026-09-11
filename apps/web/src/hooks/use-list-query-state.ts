"use client";

import { useCallback, useMemo } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import type { ListQuery } from "@/lib/api/types";

export interface ListQueryPatch {
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortDir?: "asc" | "desc";
  search?: string;
  /** Screen-specific filters (e.g. is_active); undefined/"" removes the key. */
  filters?: Record<string, string | undefined>;
}

export interface ListQueryState {
  page: number;
  pageSize: number;
  sortBy: string | undefined;
  sortDir: "asc" | "desc" | undefined;
  search: string;
  filters: Record<string, string>;
  /** The API-shaped query (shared keys only) — also the stable query-key input. */
  query: ListQuery;
  /** Merge a partial update into the URL; page resets to 1 unless set explicitly. */
  set: (patch: ListQueryPatch) => void;
}

const DEFAULT_PAGE_SIZE = 20;

interface Options {
  /** Extra URL keys this screen owns (e.g. ["is_active"]). */
  filterKeys?: string[];
}

/**
 * URL is the source of truth for list state (page/page_size/sort_by/sort_dir/
 * search + per-screen filters), so back/forward and shareable links work.
 * Every admin list screen (P03–P09) reads its state through this hook rather
 * than re-deriving it.
 */
export function useListQueryState({ filterKeys = [] }: Options = {}): ListQueryState {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const filterKeyList = filterKeys.join(",");

  const state = useMemo(() => {
    const page = Math.max(1, Number(searchParams.get("page")) || 1);
    const pageSize = Number(searchParams.get("page_size")) || DEFAULT_PAGE_SIZE;
    const sortBy = searchParams.get("sort_by") ?? undefined;
    const dirRaw = searchParams.get("sort_dir");
    const sortDir: "asc" | "desc" | undefined = dirRaw === "asc" || dirRaw === "desc" ? dirRaw : undefined;
    const search = searchParams.get("search") ?? "";
    const filters: Record<string, string> = {};
    for (const key of filterKeyList ? filterKeyList.split(",") : []) {
      const v = searchParams.get(key);
      if (v) filters[key] = v;
    }
    const query: ListQuery = { page, page_size: pageSize, sort_by: sortBy, sort_dir: sortDir, search: search || undefined };
    return { page, pageSize, sortBy, sortDir, search, filters, query };
  }, [searchParams, filterKeyList]);

  const set = useCallback<ListQueryState["set"]>(
    (patch) => {
      const next = new URLSearchParams(searchParams.toString());
      const write = (key: string, value: string | number | undefined) => {
        if (value === undefined || value === "") next.delete(key);
        else next.set(key, String(value));
      };
      if ("search" in patch) write("search", patch.search);
      if ("sortBy" in patch) write("sort_by", patch.sortBy);
      if ("sortDir" in patch) write("sort_dir", patch.sortDir);
      if ("pageSize" in patch) write("page_size", patch.pageSize);
      for (const [key, value] of Object.entries(patch.filters ?? {})) write(key, value);

      // Changing what is shown invalidates the current page number.
      const resetsPage = "search" in patch || "sortBy" in patch || "sortDir" in patch || "pageSize" in patch || "filters" in patch;
      if ("page" in patch) write("page", patch.page && patch.page > 1 ? patch.page : undefined);
      else if (resetsPage) next.delete("page");

      const qs = next.toString();
      router.replace(qs ? `${pathname}?${qs}` : pathname, { scroll: false });
    },
    [pathname, router, searchParams],
  );

  return { ...state, set };
}
