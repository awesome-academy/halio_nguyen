import { CommentRow } from "./comment-row";
import type { CommentNode } from "@/types/review.types";

interface CommentThreadNodeProps {
  reviewId: string;
  comment: CommentNode;
  depth: number;
}

// Depth is not bounded by the schema (dangling parent_id promotes to root
// server-side, but real reply chains can still run deep) — cap the visual
// indent so a deep thread doesn't push rows off-screen.
const MAX_INDENT_DEPTH = 6;
const INDENT_PX = 24;

/** Recurses over `replies` and owns indentation only — presentation lives in comment-row.tsx. */
export function CommentThreadNode({ reviewId, comment, depth }: CommentThreadNodeProps) {
  const indent = Math.min(depth, MAX_INDENT_DEPTH) * INDENT_PX;

  return (
    <div style={{ marginLeft: indent }} className="space-y-3">
      <CommentRow reviewId={reviewId} comment={comment} />
      {comment.replies.map((reply) => (
        <CommentThreadNode key={reply.id} reviewId={reviewId} comment={reply} depth={depth + 1} />
      ))}
    </div>
  );
}
