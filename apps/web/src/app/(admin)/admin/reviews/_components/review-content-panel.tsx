import { MessageSquare, ThumbsUp } from "lucide-react";
import { formatDate } from "@/lib/utils";
import type { ReviewDetail } from "@/types/review.types";

interface ReviewContentPanelProps {
  review: ReviewDetail;
}

/**
 * Read-only body/author/category/counts. The body is user-authored and
 * rendered in a high-privilege admin session — plain text only
 * (`whitespace-pre-wrap` + `break-words`), with no raw-HTML injection and
 * no markdown-to-HTML pass.
 */
export function ReviewContentPanel({ review }: ReviewContentPanelProps) {
  return (
    <section className="rounded-lg border p-5 space-y-4">
      <dl className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-4">
        <div>
          <dt className="text-muted-foreground">Author</dt>
          <dd className="font-medium">{review.author_name}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Category</dt>
          <dd className="font-medium">{review.category_name}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Created</dt>
          <dd className="font-medium">{formatDate(review.created_at)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Engagement</dt>
          <dd className="font-medium flex items-center gap-3">
            <span className="flex items-center gap-1">
              <ThumbsUp className="h-3.5 w-3.5" /> {review.like_count}
            </span>
            <span className="flex items-center gap-1">
              <MessageSquare className="h-3.5 w-3.5" /> {review.comment_count}
            </span>
          </dd>
        </div>
      </dl>

      <div className="whitespace-pre-wrap break-words text-sm leading-relaxed">{review.content}</div>
    </section>
  );
}
