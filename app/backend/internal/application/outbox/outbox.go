package outbox

import "context"

// Outbox stores entries of remaining work. A use case adds the entry in the same unit of
// work as its repository writes, so the entry commits together with them; the steps run
// after the commit (ADR-0016). It is implemented in the Infrastructure layer.
type Outbox interface {
	Add(ctx context.Context, entry Entry) error
}
