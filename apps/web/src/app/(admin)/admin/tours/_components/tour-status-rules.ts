import type { ScheduleStatus, TourStatus } from "@/types/tour.types";

export type TourAction = "publish" | "archive" | "reactivate" | "edit" | "delete";
export type ScheduleAction = "close" | "reopen" | "cancel" | "edit";

/**
 * DEC-001 — pure render decision keyed on tours.status. Shared by the list
 * row menu (tour-row-actions.tsx) and the form header (tour-status-actions.tsx)
 * so the allowed-action table is defined exactly once (Key Insight).
 */
export function allowedTourActions(status: TourStatus): TourAction[] {
  switch (status) {
    case "draft":
      return ["publish", "edit", "delete"];
    case "published":
      return ["archive", "edit"];
    case "archived":
      return ["reactivate"];
    default:
      return [];
  }
}

/**
 * DEC-002 — pure render decision keyed on tour_schedules.status. `cancelled`
 * is terminal and offers nothing (SM-002/BR-010).
 */
export function allowedScheduleActions(status: ScheduleStatus): ScheduleAction[] {
  switch (status) {
    case "open":
      return ["close", "cancel", "edit"];
    case "closed":
      return ["reopen", "cancel", "edit"];
    case "cancelled":
      return [];
    default:
      return [];
  }
}
