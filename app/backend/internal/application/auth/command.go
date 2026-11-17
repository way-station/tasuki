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

// StartLineLoginCommand starts LINE Login, from an invite link or without one. It is
// created by StartLineLoginCommandFactory.
type StartLineLoginCommand struct {
	inviteToken *string
}

// InviteToken returns the token of the invite link the user came from, and false if the
// user did not come from one.
func (command StartLineLoginCommand) InviteToken() (string, bool) {
	if command.inviteToken == nil {
		return "", false
	}
	return *command.inviteToken, true
}

// StartLineLoginCommandFactory assembles StartLineLoginCommand values.
type StartLineLoginCommandFactory struct{}

// Create returns the command. inviteToken is nil when the user did not come from an invite
// link; any other value, blank too, is looked up as a token.
func (StartLineLoginCommandFactory) Create(inviteToken *string) StartLineLoginCommand {
	return StartLineLoginCommand{inviteToken: inviteToken}
}

// CompleteLineLoginCommand is LINE's callback: the state the login started with, and the
// authorization code, or the error LINE set when the login failed or was cancelled. It is
// created by CompleteLineLoginCommandFactory.
type CompleteLineLoginCommand struct {
	state     string
	code      string
	lineError string
}

// CompleteLineLoginCommandFactory assembles CompleteLineLoginCommand values.
type CompleteLineLoginCommandFactory struct{}

// Create returns the command. code and lineError are empty when LINE did not set them.
func (CompleteLineLoginCommandFactory) Create(state, code, lineError string) CompleteLineLoginCommand {
	return CompleteLineLoginCommand{state: state, code: code, lineError: lineError}
}
