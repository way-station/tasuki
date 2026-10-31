package auth

import (
	"strings"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LoginCommand is a verified identity to log in as: a LINE user and the display name LINE
// returned. The dev-only login builds it from the request body (ADR-0009); LINE Login will
// build it from the verified ID token (ADR-0007). It is created by LoginCommandFactory.
type LoginCommand struct {
	lineUserID  string
	displayName string
}

// LoginCommandFactory assembles LoginCommand values.
type LoginCommandFactory struct{}

// Create returns the command, or a shared.ValidationError when a field is blank.
func (LoginCommandFactory) Create(lineUserID, displayName string) (LoginCommand, error) {
	lineUserID = strings.TrimSpace(lineUserID)
	displayName = strings.TrimSpace(displayName)
	if lineUserID == "" {
		return LoginCommand{}, shared.NewValidationError("line_user_id", "is required")
	}
	if displayName == "" {
		return LoginCommand{}, shared.NewValidationError("display_name", "is required")
	}
	return LoginCommand{lineUserID: lineUserID, displayName: displayName}, nil
}
