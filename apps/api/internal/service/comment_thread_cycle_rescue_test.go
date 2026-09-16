package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// Cycle-rescue cases for comment_thread_cycle_rescue.go. They share the
// newComment fixture helper declared in comment_thread_builder_test.go.

// R7 — a two-node parent_id cycle. Before promoteUnreached existed, neither
// member was a root and the root walk never entered the ring, so both
// comments silently disappeared from the response: invisible, and therefore
// unmoderatable. Both must survive, and the call must terminate.
func TestBuildCommentTree_ParentIDCyclePromotesInsteadOfVanishing(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	now := time.Now()
	tree := BuildCommentTree([]domain.Comment{
		newComment(a, &b, false, now),
		newComment(b, &a, false, now.Add(time.Minute)),
	})

	seen := map[uuid.UUID]bool{}
	var walk func(nodes []domain.CommentNode)
	walk = func(nodes []domain.CommentNode) {
		for _, n := range nodes {
			if seen[n.ID] {
				t.Fatalf("comment %s rendered twice — the cycle was not broken", n.ID)
			}
			seen[n.ID] = true
			walk(n.Replies)
		}
	}
	walk(tree)

	if !seen[a] || !seen[b] {
		t.Fatalf("both members of a parent_id cycle must stay visible, got %+v", tree)
	}
}

// R7 — a comment that is its own parent. Same failure mode as the two-node
// cycle, and the self-edge must not make the walk loop forever.
func TestBuildCommentTree_SelfParentedCommentIsPromotedToRoot(t *testing.T) {
	id := uuid.New()
	self := id
	tree := BuildCommentTree([]domain.Comment{newComment(id, &self, false, time.Now())})

	if len(tree) != 1 || tree[0].ID != id {
		t.Fatalf("a self-parented comment must be promoted to root, got %+v", tree)
	}
	if len(tree[0].Replies) != 0 {
		t.Fatalf("the self-edge must be broken, not rendered as a reply: %+v", tree[0].Replies)
	}
}

// R7 — a cycle hanging off nothing must not drag down the healthy thread
// beside it: the normal root keeps its reply, and the ring is still rescued.
func TestBuildCommentTree_CycleDoesNotDisturbHealthyThread(t *testing.T) {
	root, child := uuid.New(), uuid.New()
	a, b := uuid.New(), uuid.New()
	now := time.Now()
	tree := BuildCommentTree([]domain.Comment{
		newComment(root, nil, false, now),
		newComment(child, &root, false, now.Add(time.Minute)),
		newComment(a, &b, false, now.Add(2*time.Minute)),
		newComment(b, &a, false, now.Add(3*time.Minute)),
	})

	var healthy *domain.CommentNode
	for i := range tree {
		if tree[i].ID == root {
			healthy = &tree[i]
		}
	}
	if healthy == nil {
		t.Fatalf("the healthy root must survive alongside the cycle, got %+v", tree)
	}
	if len(healthy.Replies) != 1 || healthy.Replies[0].ID != child {
		t.Fatalf("the healthy thread must keep its reply, got %+v", healthy.Replies)
	}
}
