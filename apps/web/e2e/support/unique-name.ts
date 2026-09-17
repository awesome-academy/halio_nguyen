import { randomUUID } from "node:crypto";

/**
 * Fixed literal prefix so orphans are greppable in SQL — decisions.md §J8.
 * `categories.name` is a plain (non-partial) UNIQUE constraint and rows are
 * soft-deleted, so a name is never safe to reuse even after "delete" — the
 * random suffix is what keeps every call collision-free, not just across
 * parallel workers.
 */
export const E2E_PREFIX = "E2E_";

/** Builds a name guaranteed unique across workers and across soft-deleted history. */
export function uniqueName(label: string): string {
  return `${E2E_PREFIX}${label}_${randomUUID().slice(0, 8)}`;
}
