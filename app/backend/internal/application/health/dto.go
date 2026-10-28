package health

// StatusDTO is the result of a health check. It is created by StatusDTOFactory.
type StatusDTO struct {
	mongoReachable    bool
	postgresReachable bool
}

// Healthy reports whether every store is reachable.
func (dto StatusDTO) Healthy() bool {
	return dto.mongoReachable && dto.postgresReachable
}

// MongoReachable reports whether MongoDB is reachable.
func (dto StatusDTO) MongoReachable() bool {
	return dto.mongoReachable
}

// PostgresReachable reports whether PostgreSQL is reachable.
func (dto StatusDTO) PostgresReachable() bool {
	return dto.postgresReachable
}
