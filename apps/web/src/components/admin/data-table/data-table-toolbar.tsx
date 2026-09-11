"use client";

import { useEffect, useState } from "react";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";

interface DataTableToolbarProps {
  search: string;
  onSearchChange: (search: string) => void;
  placeholder?: string;
  /** Filter slot (selects, toggles) rendered next to the search box. */
  children?: React.ReactNode;
}

const DEBOUNCE_MS = 300;

/** Debounced search input plus a slot for per-screen filters. */
export function DataTableToolbar({ search, onSearchChange, placeholder = "Search…", children }: DataTableToolbarProps) {
  const [draft, setDraft] = useState(search);

  // Keep the box in sync when the URL changes from outside (back/forward).
  useEffect(() => setDraft(search), [search]);

  useEffect(() => {
    if (draft === search) return;
    const handle = setTimeout(() => onSearchChange(draft), DEBOUNCE_MS);
    return () => clearTimeout(handle);
  }, [draft, search, onSearchChange]);

  return (
    <div className="flex items-center gap-3 mb-4">
      <div className="relative w-72">
        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input value={draft} onChange={(e) => setDraft(e.target.value)} placeholder={placeholder} className="pl-9" aria-label="Search" />
      </div>
      {children}
    </div>
  );
}
