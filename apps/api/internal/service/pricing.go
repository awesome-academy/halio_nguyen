package service

// EffectivePrice implements ALG-001 — the price actually charged for a given
// departure: the schedule's own price_override wins if set, otherwise the
// tour's discount_price if set, otherwise the tour's base price. Pure
// arithmetic, no DB access, so it is unit-testable in isolation (phase-04
// Implementation Steps #1).
func EffectivePrice(priceOverride, discountPrice *float64, price float64) float64 {
	if priceOverride != nil {
		return *priceOverride
	}
	if discountPrice != nil {
		return *discountPrice
	}
	return price
}
