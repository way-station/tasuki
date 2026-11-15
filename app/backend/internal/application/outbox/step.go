// Package outbox records the work that remains after a use case writes to PostgreSQL, so
// that the writes to MongoDB and the final PostgreSQL step are completed even when they
// fail the first time (ADR-0016).
package outbox

// Step is one step of the remaining work. Each feature defines its own steps in its
// package, for example the running order history in the running order layer. The
// Infrastructure layer runs each kind, so a step holds only data that can be stored, and
// running it twice must have the same effect as running it once (ADR-0016).
type Step interface {
	// Kind names the step, for example "runningorder.append_history". It tells the
	// Infrastructure layer how to run the step.
	Kind() string
}
