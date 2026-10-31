package presentation

import (
	"context"
	"errors"

	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
)

// errNotImplemented is returned for operations of the OpenAPI document that have no
// controller yet. The response error handler answers it with 501.
var errNotImplemented = errors.New("not implemented")

// notImplemented has every operation of api.StrictServerInterface and answers each with
// errNotImplemented. Server embeds it one level deeper than the controllers, so an
// operation a controller has is served by the controller. When the document gains an
// operation, the build fails until it is added here.
type notImplemented struct{}

func (notImplemented) AuthorizeLine(context.Context, api.AuthorizeLineRequestObject) (api.AuthorizeLineResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) LineCallback(context.Context, api.LineCallbackRequestObject) (api.LineCallbackResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) DevLogin(context.Context, api.DevLoginRequestObject) (api.DevLoginResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) GetInviteLink(context.Context, api.GetInviteLinkRequestObject) (api.GetInviteLinkResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) GetMe(context.Context, api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) GetMyProfile(context.Context, api.GetMyProfileRequestObject) (api.GetMyProfileResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) ListPendingMemberLinks(context.Context, api.ListPendingMemberLinksRequestObject) (api.ListPendingMemberLinksResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) ApproveMemberLink(context.Context, api.ApproveMemberLinkRequestObject) (api.ApproveMemberLinkResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) RejectMemberLink(context.Context, api.RejectMemberLinkRequestObject) (api.RejectMemberLinkResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) ListRaces(context.Context, api.ListRacesRequestObject) (api.ListRacesResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) CreateRace(context.Context, api.CreateRaceRequestObject) (api.CreateRaceResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) GetRace(context.Context, api.GetRaceRequestObject) (api.GetRaceResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) UpdateRace(context.Context, api.UpdateRaceRequestObject) (api.UpdateRaceResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) AnswerAttendance(context.Context, api.AnswerAttendanceRequestObject) (api.AnswerAttendanceResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) ListEntries(context.Context, api.ListEntriesRequestObject) (api.ListEntriesResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) CreateEntry(context.Context, api.CreateEntryRequestObject) (api.CreateEntryResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) RemoveEntryMember(context.Context, api.RemoveEntryMemberRequestObject) (api.RemoveEntryMemberResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) AddEntryMember(context.Context, api.AddEntryMemberRequestObject) (api.AddEntryMemberResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) SaveRunningOrder(context.Context, api.SaveRunningOrderRequestObject) (api.SaveRunningOrderResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) ConfirmRunningOrder(context.Context, api.ConfirmRunningOrderRequestObject) (api.ConfirmRunningOrderResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) ListInvitations(context.Context, api.ListInvitationsRequestObject) (api.ListInvitationsResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) CancelInvitation(context.Context, api.CancelInvitationRequestObject) (api.CancelInvitationResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) Invite(context.Context, api.InviteRequestObject) (api.InviteResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) ListRaceLaps(context.Context, api.ListRaceLapsRequestObject) (api.ListRaceLapsResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) GetMyLaps(context.Context, api.GetMyLapsRequestObject) (api.GetMyLapsResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) SubmitLaps(context.Context, api.SubmitLapsRequestObject) (api.SubmitLapsResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) DeleteLap(context.Context, api.DeleteLapRequestObject) (api.DeleteLapResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) EditLap(context.Context, api.EditLapRequestObject) (api.EditLapResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) ListRunningOrders(context.Context, api.ListRunningOrdersRequestObject) (api.ListRunningOrdersResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) ListRosterMembers(context.Context, api.ListRosterMembersRequestObject) (api.ListRosterMembersResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) CreateRosterMember(context.Context, api.CreateRosterMemberRequestObject) (api.CreateRosterMemberResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) GetRosterMember(context.Context, api.GetRosterMemberRequestObject) (api.GetRosterMemberResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) UpdateRosterMember(context.Context, api.UpdateRosterMemberRequestObject) (api.UpdateRosterMemberResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) CreateInviteLink(context.Context, api.CreateInviteLinkRequestObject) (api.CreateInviteLinkResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) UnlinkRosterMember(context.Context, api.UnlinkRosterMemberRequestObject) (api.UnlinkRosterMemberResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) Logout(context.Context, api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) CreateSession(context.Context, api.CreateSessionRequestObject) (api.CreateSessionResponseObject, error) {
	return nil, errNotImplemented
}

func (notImplemented) GetHealth(context.Context, api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return nil, errNotImplemented
}
