// Package service holds business-rule logic that sits between handlers and
// repositories: transactions, cross-repository guards, and validation that
// depends on more than one table.
//
// It is intentionally empty in this phase. No cross-cutting service-layer
// helper is needed yet — YAGNI. Later phases add their own
// feature-scoped services here (or in their own files within this package)
// as soon as a second consumer actually needs the shared logic.
package service
