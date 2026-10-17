package health

// StatusDTOFactory assembles StatusDTO values.
type StatusDTOFactory struct{}

// Create returns the status for the given store states.
func (StatusDTOFactory) Create(mongoReachable bool) StatusDTO {
	return StatusDTO{mongoReachable: mongoReachable}
}
