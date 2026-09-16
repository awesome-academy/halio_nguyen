import { Badge } from "@/components/ui/badge";
import type { ReviewStatus } from "@/types/review.types";

const STATUS_VARIANT: Record<ReviewStatus, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Draft", variant: "outline" },
  published: { label: "Published", variant: "default" },
  hidden: { label: "Hidden", variant: "secondary" },
};

interface ReviewStatusBadgeProps {
  status: ReviewStatus;
  className?: string;
}

/** Shared draft/published/hidden badge for the reviews list and detail screens. */
export function ReviewStatusBadge({ status, className }: ReviewStatusBadgeProps) {
  const s = STATUS_VARIANT[status];
  return (
    <Badge variant={s.variant} className={className}>
      {s.label}
    </Badge>
  );
}
