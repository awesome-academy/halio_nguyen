package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// newComment builds a flat domain.Comment fixture. deleted sets DeletedAt to
// a time after createdAt, mirroring a real soft-delete.
func newComment(id uuid.UUID, parent *uuid.UUID, deleted bool, createdAt time.Time) domain.Comment {
	c := domain.Comment{
		ID:        id,
		ParentID:  parent,
		Content:   "content-" + id.String()[:8],
		CreatedAt: createdAt,
		User:      &domain.User{ID: uuid.New(), FullName: "Author"},
	}
	if deleted {
		deletedAt := createdAt.Add(time.Hour)
		c.DeletedAt = &deletedAt
	}
	return c
}

// R1/BR-003 — the phase's sharpest case: a soft-deleted parent's live
// replies must stay visible and moderable beneath a content-free placeholder.
func TestBuildCommentTree_DeletedParentWithLiveRepliesKeepsPlaceholderAndReplies(t *testing.T) {
	base := time.Now()
	parent, reply1, reply2 := uuid.New(), uuid.New(), uuid.New()

	comments := []domain.Comment{
		newComment(parent, nil, true, base),
		newComment(reply1, &parent, false, base.Add(time.Minute)),
		newComment(reply2, &parent, false, base.Add(2*time.Minute)),
	}

	tree := BuildCommentTree(comments)
	if len(tree) != 1 {
		t.Fatalf("want 1 root, got %d: %+v", len(tree), tree)
	}
	root := tree[0]
	if !root.IsDeleted {
		t.Error("deleted parent must render as a placeholder (is_deleted: true)")
	}
	if root.Content != "" {
		t.Errorf("deleted parent must not leak content, got %q", root.Content)
	}
	if root.User != nil {
		t.Errorf("deleted parent must not leak author, got %+v", root.User)
	}
	if len(root.Replies) != 2 {
		t.Fatalf("want 2 live replies beneath the placeholder, got %d", len(root.Replies))
	}
	if root.Replies[0].ID != reply1 || root.Replies[1].ID != reply2 {
		t.Error("replies must preserve created_at ASC order")
	}
	if root.Replies[0].IsDeleted || root.Replies[1].IsDeleted {
		t.Error("live replies must not be marked deleted")
	}
}

// A deleted comment with no surviving descendant must be pruned entirely —
// no phantom [deleted] row.
func TestBuildCommentTree_DeletedChildlessLeafIsPruned(t *testing.T) {
	base := time.Now()
	live, deletedLeaf := uuid.New(), uuid.New()

	tree := BuildCommentTree([]domain.Comment{
		newComment(live, nil, false, base),
		newComment(deletedLeaf, nil, true, base.Add(time.Minute)),
	})

	if len(tree) != 1 {
		t.Fatalf("want 1 surviving root (deleted leaf pruned), got %d: %+v", len(tree), tree)
	}
	if tree[0].ID != live {
		t.Errorf("surviving root = %v, want %v", tree[0].ID, live)
	}
}

// A deleted parent whose only child is also deleted has no surviving
// descendant either — both are pruned.
func TestBuildCommentTree_DeletedParentWithOnlyDeletedChildIsFullyPruned(t *testing.T) {
	base := time.Now()
	keep, parent, child := uuid.New(), uuid.New(), uuid.New()

	tree := BuildCommentTree([]domain.Comment{
		newComment(keep, nil, false, base),
		newComment(parent, nil, true, base.Add(time.Minute)),
		newComment(child, &parent, true, base.Add(2*time.Minute)),
	})

	if len(tree) != 1 || tree[0].ID != keep {
		t.Fatalf("deleted parent+child branch must be fully pruned, got %+v", tree)
	}
}

// A comment whose parent_id points at a row absent from the result set (a
// soft-deleted parent fetched by a different, narrower query, or any other
// break in the chain) must be promoted to root, never dropped.
func TestBuildCommentTree_DanglingParentIsPromotedToRoot(t *testing.T) {
	orphan, missingParent := uuid.New(), uuid.New()

	tree := BuildCommentTree([]domain.Comment{
		newComment(orphan, &missingParent, false, time.Now()),
	})

	if len(tree) != 1 || tree[0].ID != orphan {
		t.Fatalf("comment with a dangling parent_id must be promoted to root, got %+v", tree)
	}
}

// Depth is not bounded by the schema — a 200-deep reply chain must assemble
// without a recursive Go function blowing the stack (R7).
func TestBuildCommentTree_DeepChainAssemblesWithoutRecursion(t *testing.T) {
	const depth = 200
	base := time.Now()
	comments := make([]domain.Comment, 0, depth)
	ids := make([]uuid.UUID, depth)
	var parent *uuid.UUID
	for i := 0; i < depth; i++ {
		id := uuid.New()
		ids[i] = id
		comments = append(comments, newComment(id, parent, false, base.Add(time.Duration(i)*time.Second)))
		parentCopy := id
		parent = &parentCopy
	}

	tree := BuildCommentTree(comments)
	if len(tree) != 1 || tree[0].ID != ids[0] {
		t.Fatalf("want a single root chain head %v, got %+v", ids[0], tree)
	}

	node := tree[0]
	for i := 1; i < depth; i++ {
		if len(node.Replies) != 1 {
			t.Fatalf("depth %d: want exactly 1 reply, got %d", i, len(node.Replies))
		}
		if node.Replies[0].ID != ids[i] {
			t.Fatalf("depth %d: reply id = %v, want %v", i, node.Replies[0].ID, ids[i])
		}
		node = node.Replies[0]
	}
}

// Empty input must return an empty, non-nil slice — []CommentNode{}, so the
// wire ships `[]` rather than `null`.
func TestBuildCommentTree_EmptyInputReturnsEmptyNonNilSlice(t *testing.T) {
	tree := BuildCommentTree(nil)
	if tree == nil {
		t.Fatal("must return a non-nil empty slice")
	}
	if len(tree) != 0 {
		t.Fatalf("want 0 roots, got %d", len(tree))
	}
}

// Replies must always serialize as [] rather than null (Constraints).
func TestBuildCommentTree_RepliesIsNeverNil(t *testing.T) {
	tree := BuildCommentTree([]domain.Comment{newComment(uuid.New(), nil, false, time.Now())})
	if tree[0].Replies == nil {
		t.Error("Replies must be a non-nil empty slice when there are no replies")
	}
}
