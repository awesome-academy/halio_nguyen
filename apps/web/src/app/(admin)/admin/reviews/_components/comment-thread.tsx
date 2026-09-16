import { CommentThreadNode } from "./comment-thread-node";
import type { CommentNode } from "@/types/review.types";

interface CommentThreadProps {
  reviewId: string;
  comments: CommentNode[];
}

/** Root of the comment tree: maps top-level nodes and renders the empty state (US002). */
export function CommentThread({ reviewId, comments }: CommentThreadProps) {
  return (
    <section className="rounded-lg border p-5 space-y-4">
      <h2 className="text-sm font-semibold">Comments</h2>
      {comments.length === 0 ? (
        <p className="text-sm text-muted-foreground">This review has no comments yet.</p>
      ) : (
        <div className="space-y-3">
          {comments.map((comment) => (
            <CommentThreadNode key={comment.id} reviewId={reviewId} comment={comment} depth={0} />
          ))}
        </div>
      )}
    </section>
  );
}
