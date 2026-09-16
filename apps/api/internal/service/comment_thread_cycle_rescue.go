package service

import (
	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// promoteUnreached rescues comments the root walk never reached. Every such
// comment sits in a parent_id cycle (A -> B -> A, or a row that is its own
// parent): each member has a parent present in the map, so none was promoted
// at attach time, and a walk that starts only from roots never enters the
// ring. Left alone they would vanish from the response entirely — invisible
// and therefore unmoderatable, the one thing this builder must never do.
//
// Each unreached comment is detached from its parent before becoming a root,
// which is what breaks the ring: without that, the BFS below would follow
// the cycle forever. Walking from the new root immediately marks the rest of
// its component reached, so a ring degrades to a flat list plus its intact
// sub-threads rather than disappearing (R7).
//
// Nothing in this application re-parents a comment, so this is defence
// against a direct database edit, not against a code path.
func promoteUnreached(
	comments []domain.Comment, roots, order []uuid.UUID, slots map[uuid.UUID]*commentSlot,
) (newRoots, newOrder []uuid.UUID) {
	reached := make(map[uuid.UUID]bool, len(slots))
	for _, id := range order {
		reached[id] = true
	}

	for _, c := range comments {
		if reached[c.ID] {
			continue
		}
		if c.ParentID != nil {
			if parent, ok := slots[*c.ParentID]; ok {
				parent.children = removeID(parent.children, c.ID)
			}
		}
		roots = append(roots, c.ID)
		for _, id := range breadthFirstOrder([]uuid.UUID{c.ID}, slots) {
			reached[id] = true
			order = append(order, id)
		}
	}
	return roots, order
}

// removeID drops the first occurrence of id, preserving sibling order.
func removeID(ids []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for i, candidate := range ids {
		if candidate == id {
			return append(ids[:i:i], ids[i+1:]...)
		}
	}
	return ids
}
