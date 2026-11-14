package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/url"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
)

// Errors that LINE Login's callback hands to the frontend in error= (docs/design/api.md).
const (
	callbackInviteUnusable = "invite_unusable"
	callbackLeft           = "left"
	callbackLoginFailed    = "login_failed"
)

// LineLoginController serves LINE Login with a one-time code (ADR-0007, ADR-0014): checking
// an invite link, starting the login, LINE's callback, and exchanging the code for a session.
type LineLoginController struct {
	checkInvite     *auth.CheckInviteLinkUseCase
	start           *auth.StartLineLoginUseCase
	complete        *auth.CompleteLineLoginUseCase
	exchange        *auth.ExchangeLoginCodeUseCase
	startFactory    auth.StartLineLoginCommandFactory
	completeFactory auth.CompleteLineLoginCommandFactory
	frontendOrigin  string
	logger          *slog.Logger
}

// NewLineLoginController returns a LineLoginController. The callback redirects to
// frontendOrigin (FRONTEND_ORIGIN) followed by /auth/complete.
func NewLineLoginController(checkInvite *auth.CheckInviteLinkUseCase, start *auth.StartLineLoginUseCase, complete *auth.CompleteLineLoginUseCase, exchange *auth.ExchangeLoginCodeUseCase, startFactory auth.StartLineLoginCommandFactory, completeFactory auth.CompleteLineLoginCommandFactory, frontendOrigin string, logger *slog.Logger) *LineLoginController {
	return &LineLoginController{checkInvite: checkInvite, start: start, complete: complete, exchange: exchange, startFactory: startFactory, completeFactory: completeFactory, frontendOrigin: frontendOrigin, logger: logger}
}

// GetInviteLink returns the team name of a usable invite link. An unusable one is 404
// invite_unusable.
func (controller *LineLoginController) GetInviteLink(ctx context.Context, request api.GetInviteLinkRequestObject) (api.GetInviteLinkResponseObject, error) {
	link, err := controller.checkInvite.Execute(ctx, request.Token)
	if errors.Is(err, identity.ErrInviteLinkUnusable) {
		return api.GetInviteLink404ApplicationProblemPlusJSONResponse(inviteUnusable()), nil
	}
	if err != nil {
		return nil, err
	}
	return api.GetInviteLink200JSONResponse{TeamName: link.TeamName()}, nil
}

// AuthorizeLine starts LINE Login and returns the URL of LINE's login page. An unusable
// invite token is 404 invite_unusable.
func (controller *LineLoginController) AuthorizeLine(ctx context.Context, request api.AuthorizeLineRequestObject) (api.AuthorizeLineResponseObject, error) {
	if request.Body == nil {
		return api.AuthorizeLine400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	authorizationURL, err := controller.start.Execute(ctx, controller.startFactory.Create(request.Body.InviteToken))
	if errors.Is(err, identity.ErrInviteLinkUnusable) {
		return api.AuthorizeLine404ApplicationProblemPlusJSONResponse(inviteUnusable()), nil
	}
	if err != nil {
		return nil, err
	}
	return api.AuthorizeLine200JSONResponse{AuthorizationUrl: authorizationURL}, nil
}

// LineCallback finishes LINE Login and redirects to the frontend's /auth/complete with the
// one-time code, or with error=invite_unusable, left, or login_failed. Any other failure is
// logged and redirects with login_failed, so the browser never stops on an error page of the
// backend.
func (controller *LineLoginController) LineCallback(ctx context.Context, request api.LineCallbackRequestObject) (api.LineCallbackResponseObject, error) {
	code, err := controller.complete.Execute(ctx, controller.completeFactory.Create(request.Params.State, valueOf(request.Params.Code), valueOf(request.Params.Error)))
	query := url.Values{}
	switch {
	case err == nil:
		query.Set("code", code)
	case errors.Is(err, identity.ErrInviteLinkUnusable):
		query.Set("error", callbackInviteUnusable)
	case errors.Is(err, auth.ErrMemberLeft):
		query.Set("error", callbackLeft)
	case errors.Is(err, auth.ErrLoginFailed):
		controller.logger.InfoContext(ctx, "LINE Login failed", "error", err)
		query.Set("error", callbackLoginFailed)
	default:
		controller.logger.ErrorContext(ctx, "LINE Login callback failed", "error", err)
		query.Set("error", callbackLoginFailed)
	}
	return api.LineCallback302Response{Headers: api.LineCallback302ResponseHeaders{Location: controller.frontendOrigin + "/auth/complete?" + query.Encode()}}, nil
}

// CreateSession exchanges the one-time code for a session and sets the session cookie. The
// session token is not in the body. An unknown, expired, or used code is 400.
func (controller *LineLoginController) CreateSession(ctx context.Context, request api.CreateSessionRequestObject) (api.CreateSessionResponseObject, error) {
	if request.Body == nil {
		return api.CreateSession400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	session, err := controller.exchange.Execute(ctx, request.Body.Code)
	if errors.Is(err, auth.ErrInvalidLoginCode) {
		return api.CreateSession400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("the code is unknown, expired, or already used")}, nil
	}
	if err != nil {
		return nil, err
	}
	cookie := sessionCookie(session.Token(), session.ExpiresAt())
	return api.CreateSession200JSONResponse{
		Body:    api.Session{ExpiresAt: session.ExpiresAt()},
		Headers: api.CreateSession200ResponseHeaders{SetCookie: &cookie},
	}, nil
}

func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
