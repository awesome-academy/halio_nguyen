"use client";

import Link from "next/link";
import { Archive, CheckCircle2, MoreHorizontal, Pencil, RotateCcw, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import type { TourListItem } from "@/types/tour.types";
import { allowedTourActions } from "./tour-status-rules";

export interface TourRowActionHandlers {
  onPublish: (tour: TourListItem) => void;
  onArchive: (tour: TourListItem) => void;
  onReactivate: (tour: TourListItem) => void;
  onDelete: (tour: TourListItem) => void;
}

interface TourRowActionsProps extends TourRowActionHandlers {
  tour: TourListItem;
}

/** DEC-001's list-row menu — the offered items come straight from allowedTourActions. */
export function TourRowActions({ tour, onPublish, onArchive, onReactivate, onDelete }: TourRowActionsProps) {
  const actions = allowedTourActions(tour.status);

  return (
    <div className="flex justify-end">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" className="h-8 w-8" aria-label={`Actions for ${tour.title}`}>
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {actions.includes("edit") && (
            <DropdownMenuItem asChild>
              <Link href={`/admin/tours/${tour.id}`}>
                <Pencil className="h-4 w-4 mr-2" /> Edit
              </Link>
            </DropdownMenuItem>
          )}
          {actions.includes("publish") && (
            <DropdownMenuItem onClick={() => onPublish(tour)}>
              <CheckCircle2 className="h-4 w-4 mr-2" /> Publish
            </DropdownMenuItem>
          )}
          {actions.includes("archive") && (
            <DropdownMenuItem onClick={() => onArchive(tour)}>
              <Archive className="h-4 w-4 mr-2" /> Archive
            </DropdownMenuItem>
          )}
          {actions.includes("reactivate") && (
            <DropdownMenuItem onClick={() => onReactivate(tour)}>
              <RotateCcw className="h-4 w-4 mr-2" /> Reactivate
            </DropdownMenuItem>
          )}
          {actions.includes("delete") && (
            <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onDelete(tour)}>
              <Trash2 className="h-4 w-4 mr-2" /> Delete
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
