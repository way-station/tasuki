package controller

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

// AuthController serves the auth and joining operations of the OpenAPI document that are
// built so far: the dev-only login, logout, and GET /me.
type AuthController struct {
	login          *auth.LoginUseCase
	logout         *auth.LogoutUseCase
	getMe          *auth.GetMeUseCase
	commandFactory auth.LoginCommandFactory
}

// NewAuthController returns an AuthController.
func NewAuthController(login *auth.LoginUseCase, logout *auth.LogoutUseCase, getMe *auth.GetMeUseCase, commandFactory auth.LoginCommandFactory) *AuthController {
	return &AuthController{login: login, logout: logout, getMe: getMe, commandFactory: commandFactory}
}

// DevLogin logs in as any LINE user without LINE (ADR-0009). The route exists only when
// DEV_LOGIN_ENABLED is set.
func (controller *AuthController) DevLogin(ctx context.Context, request api.DevLoginRequestObject) (api.DevLoginResponseObject, error) {
	if request.Body == nil {
		return api.DevLogin400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	command, err := controller.commandFactory.Create(request.Body.LineUserId, request.Body.DisplayName)
	var validationErr *shared.ValidationError
	if errors.As(err, &validationErr) {
		return api.DevLogin400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(validationErr.Error())}, nil
	}
	if err != nil {
		return nil, err
	}

	session, err := controller.login.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	cookie := sessionCookie(session.Token(), session.ExpiresAt())
	return api.DevLogin200JSONResponse{
		Body:    api.Session{ExpiresAt: session.ExpiresAt()},
		Headers: api.DevLogin200ResponseHeaders{SetCookie: &cookie},
	}, nil
}

// Logout revokes the caller's session and clears the cookie.
func (controller *AuthController) Logout(ctx context.Context, _ api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	token, ok := middleware.SessionToken(ctx)
	if !ok {
		return api.Logout401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: api.UnauthenticatedApplicationProblemPlusJSONResponse(middleware.UnauthenticatedProblem())}, nil
	}
	if err := controller.logout.Execute(ctx, token); err != nil {
		return nil, err
	}
	cookie := clearedSessionCookie()
	return api.Logout204Response{Headers: api.Logout204ResponseHeaders{SetCookie: &cookie}}, nil
}

// GetMe returns who the caller is and which screens they may use.
func (controller *AuthController) GetMe(ctx context.Context, _ api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	unauthenticated := api.GetMe401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: api.UnauthenticatedApplicationProblemPlusJSONResponse(middleware.UnauthenticatedProblem())}
	userID, ok := middleware.UserID(ctx)
	if !ok {
		return unauthenticated, nil
	}
	caller, err := controller.getMe.Execute(ctx, userID)
	if errors.Is(err, auth.ErrUnauthenticated) {
		return unauthenticated, nil
	}
	if err != nil {
		return nil, err
	}

	response := api.GetMe200JSONResponse{
		DisplayName: caller.DisplayName(),
		IsOrganizer: caller.IsOrganizer(),
		Status:      meStatus(caller.Status()),
	}
	if member, ok := caller.RosterMember(); ok {
		response.RosterMember = &api.RosterMemberBasic{Id: string(member.ID()), Name: member.Name()}
	}
	return response, nil
}

func meStatus(status identity.CallerStatus) api.MeStatus {
	switch status {
	case identity.CallerApproved:
		return api.MeStatusApproved
	case identity.CallerPending:
		return api.MeStatusPending
	case identity.CallerRejected:
		return api.MeStatusRejected
	default:
		return api.MeStatusNone
	}
}

func badRequest(detail string) api.BadRequestApplicationProblemPlusJSONResponse {
	return api.BadRequestApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Bad Request", Status: http.StatusBadRequest, Code: api.ProblemCodeInvalidRequest, Detail: &detail}
}

// sessionCookie is the Set-Cookie value for a new session (ADR-0014).
func sessionCookie(token string, expiresAt time.Time) string {
	cookie := &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/api",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	return cookie.String()
}

func clearedSessionCookie() string {
	cookie := &http.Cookie{
		Name:     middleware.SessionCookieName,
		Path:     "/api",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	return cookie.String()
}
