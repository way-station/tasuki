package health

// StatusDTO is the result of a health check. It is created by StatusDTOFactory.
type StatusDTO struct {
	mongoReachable    bool
	postgresReachable bool
}

// Healthy reports whether every store is reachable.
func (d StatusDTO) Healthy() bool {
	return d.mongoReachable && d.postgresReachable
}

// MongoReachable reports whether MongoDB is reachable.
func (d StatusDTO) MongoReachable() bool {
	return d.mongoReachable
}

// PostgresReachable reports whether PostgreSQL is reachable.
func (d StatusDTO) PostgresReachable() bool {
	return d.postgresReachable
}
