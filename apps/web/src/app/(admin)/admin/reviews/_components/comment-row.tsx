import { cn } from "@/lib/utils";
import { formatDate } from "@/lib/utils";
import { CommentActions } from "./comment-actions";
import type { CommentNode } from "@/types/review.types";

interface CommentRowProps {
  reviewId: string;
  comment: CommentNode;
}

/**
 * One comment's presentation in its three visual states: normal, `is_hidden`
 * (muted + "Hidden" badge), `is_deleted` (structural `[deleted]` placeholder
 * — no author, no body). Comment text is user-authored: rendered as plain
 * text (`whitespace-pre-wrap break-words`) — no raw-HTML injection and no
 * markdown pass, so nothing here can execute in the admin session.
 */
export function CommentRow({ reviewId, comment }: CommentRowProps) {
  if (comment.is_deleted) {
    return (
      <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">
        [deleted] · {formatDate(comment.created_at)}
      </div>
    );
  }

  return (
    <div className={cn("rounded-md border p-3 text-sm", comment.is_hidden && "opacity-60")}>
      <div className="flex items-center justify-between gap-2 mb-1">
        <div className="flex items-center gap-2">
          <span className="font-medium">{comment.user?.full_name}</span>
          {comment.is_hidden && (
            <span className="rounded-full bg-muted px-2 py-0.5 text-[11px] font-semibold text-muted-foreground">Hidden</span>
          )}
        </div>
        <span className="text-xs text-muted-foreground">{formatDate(comment.created_at)}</span>
      </div>
      <p className={cn("whitespace-pre-wrap break-words", comment.is_hidden && "line-through")}>{comment.content}</p>
      <div className="mt-2">
        <CommentActions reviewId={reviewId} comment={comment} />
      </div>
    </div>
  );
}
