package postgres

// sqlc generates sqlcgen/ from queries/ and the up files in migrations/ (ADR-0015).
// Run `make generate` after changing either; never edit the generated files by hand.
//go:generate sqlc generate -f sqlc.yaml
