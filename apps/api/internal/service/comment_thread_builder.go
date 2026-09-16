package service

import (
	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// BuildCommentTree assembles F006 A2's nested reply tree from a flat,
// review-scoped comment slice — every row, deleted or not (see
// ReviewAdminRepository.ListCommentsForReview's doc comment for why the
// caller must never filter deleted_at before this runs). It is a pure
// function: no DB access, no context, safe to call from any layer that
// already has the rows in hand, and the highest-value test in the phase
// (comment_thread_builder_test.go) exercises it directly.
//
// The pass runs in four stages, each a single O(n) walk over the input:
//
//  1. map every comment by id.
//  2. attach each comment under its parent; a comment whose parent_id is
//     absent from the map (a dangling reference — see the repository's
//     doc comment for why this can legitimately happen) is promoted to
//     root instead of discarded. No comment may become invisible.
//  3. prune bottom-up: a deleted comment with no surviving (kept)
//     descendant is dropped entirely rather than rendered as a phantom
//     [deleted] row.
//  4. redact: any deleted comment that survives pruning (because it has at
//     least one live descendant, BR-003) is reduced to
//     {id, parent_id, created_at, is_deleted: true} — content and author
//     never populated, so neither can leak to the wire (R8).
//
// Depth is not bounded by the schema, so both the pruning pass and the
// node-building pass walk an iteratively computed breadth-first order
// rather than recursing — a deep or even cyclic parent_id chain degrades to
// a flat list instead of overflowing the stack (R7).
func BuildCommentTree(comments []domain.Comment) []domain.CommentNode {
	slots := make(map[uuid.UUID]*commentSlot, len(comments))
	for i := range comments {
		slots[comments[i].ID] = &commentSlot{comment: comments[i]}
	}

	roots := make([]uuid.UUID, 0)
	for _, c := range comments {
		if c.ParentID != nil {
			if parent, ok := slots[*c.ParentID]; ok {
				parent.children = append(parent.children, c.ID)
				continue
			}
		}
		roots = append(roots, c.ID)
	}

	order := breadthFirstOrder(roots, slots)
	if len(order) < len(slots) {
		roots, order = promoteUnreached(comments, roots, order, slots)
	}
	keep := computeKeep(order, slots)

	built := make(map[uuid.UUID]domain.CommentNode, len(order))
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		if keep[id] {
			built[id] = buildCommentNode(slots[id], keep, built)
		}
	}

	result := make([]domain.CommentNode, 0, len(roots))
	for _, id := range roots {
		if keep[id] {
			result = append(result, built[id])
		}
	}
	return result
}

// commentSlot bundles one flat comment with its discovered children ids, in
// the order ListCommentsForReview returns them (created_at ASC).
type commentSlot struct {
	comment  domain.Comment
	children []uuid.UUID
}

// breadthFirstOrder walks the forest iteratively (a plain queue, no
// recursion) and returns every id in an order where a parent always
// precedes its children. That ordering is what lets computeKeep and
// buildCommentTree process the reversed order bottom-up: by the time a node
// is visited, every one of its children has already been visited.
func breadthFirstOrder(roots []uuid.UUID, slots map[uuid.UUID]*commentSlot) []uuid.UUID {
	order := make([]uuid.UUID, 0, len(slots))
	queue := append([]uuid.UUID(nil), roots...)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		queue = append(queue, slots[id].children...)
	}
	return order
}

// computeKeep decides, for every comment, whether it survives pruning: a
// non-deleted comment is always kept; a deleted one is kept only if at
// least one of its children is kept (BR-003 — a deleted parent with a live
// descendant must remain, as a placeholder).
func computeKeep(order []uuid.UUID, slots map[uuid.UUID]*commentSlot) map[uuid.UUID]bool {
	keep := make(map[uuid.UUID]bool, len(order))
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		slot := slots[id]

		kept := slot.comment.DeletedAt == nil
		if !kept {
			for _, childID := range slot.children {
				if keep[childID] {
					kept = true
					break
				}
			}
		}
		keep[id] = kept
	}
	return keep
}

// buildCommentNode renders one kept slot into its wire shape. built must
// already hold every one of slot's kept children — guaranteed by the
// caller's reverse-order walk.
func buildCommentNode(slot *commentSlot, keep map[uuid.UUID]bool, built map[uuid.UUID]domain.CommentNode) domain.CommentNode {
	replies := make([]domain.CommentNode, 0, len(slot.children))
	for _, childID := range slot.children {
		if keep[childID] {
			replies = append(replies, built[childID])
		}
	}

	node := domain.CommentNode{
		ID:        slot.comment.ID,
		ParentID:  slot.comment.ParentID,
		CreatedAt: slot.comment.CreatedAt,
		Replies:   replies,
	}

	if slot.comment.DeletedAt != nil {
		// Redacted: Content/IsHidden/User stay at their zero values ("",
		// false, nil) — a deleted comment's content and author must never
		// reach the wire (R8).
		node.IsDeleted = true
		return node
	}

	node.Content = slot.comment.Content
	node.IsHidden = slot.comment.IsHidden
	if slot.comment.User != nil {
		node.User = &domain.CommentAuthor{
			ID:        slot.comment.User.ID,
			FullName:  slot.comment.User.FullName,
			AvatarURL: slot.comment.User.AvatarURL,
		}
	}
	return node
}
