package controller

import (
	"context"
	"errors"

	"github.com/oapi-codegen/nullable"

	"github.com/way-station/tasuki/app/backend/internal/application/participations"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

// ParticipationUseCases are the use cases ParticipationController calls.
type ParticipationUseCases struct {
	ListInvitations   *participations.ListInvitationsUseCase
	Invite            *participations.InviteUseCase
	CancelInvitation  *participations.CancelInvitationUseCase
	AnswerAttendance  *participations.AnswerAttendanceUseCase
	ListEntries       *participations.ListEntriesUseCase
	CreateEntry       *participations.CreateEntryUseCase
	AddEntryMember    *participations.AddEntryMemberUseCase
	RemoveEntryMember *participations.RemoveEntryMemberUseCase
}

// ParticipationController serves invitations, attendance, and entries
// (docs/design/api.md). The routes check the caller's role before it runs.
type ParticipationController struct {
	useCases       ParticipationUseCases
	commandFactory participations.ParticipationCommandFactory
}

// NewParticipationController returns a ParticipationController.
func NewParticipationController(useCases ParticipationUseCases, commandFactory participations.ParticipationCommandFactory) *ParticipationController {
	return &ParticipationController{useCases: useCases, commandFactory: commandFactory}
}

// ListInvitations returns every roster member who has not left, with their invitation to
// the race.
func (controller *ParticipationController) ListInvitations(ctx context.Context, request api.ListInvitationsRequestObject) (api.ListInvitationsResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.ListInvitations401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	invitations, err := controller.useCases.ListInvitations.Execute(ctx, caller, shared.RaceID(request.RaceId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.ListInvitations404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	response := api.ListInvitations200JSONResponse{Members: make([]api.InvitationMember, 0, len(invitations))}
	for _, invitation := range invitations {
		member := api.InvitationMember{RosterMemberId: string(invitation.MemberID()), Name: invitation.Name(), Invited: invitation.Invited(), Attendance: nullable.NewNullNullable[api.AttendanceStatus]()}
		if status, answered := invitation.Attendance(); answered {
			member.Attendance = nullable.NewNullableWithValue(api.AttendanceStatus(status))
		}
		response.Members = append(response.Members, member)
	}
	return response, nil
}

// Invite invites a roster member to the race. A member who has left is 422.
func (controller *ParticipationController) Invite(ctx context.Context, request api.InviteRequestObject) (api.InviteResponseObject, error) {
	command := controller.commandFactory.Invitation(shared.RaceID(request.RaceId), shared.MemberID(request.RosterMemberId))
	err := controller.useCases.Invite.Execute(ctx, command)
	if errors.Is(err, shared.ErrNotFound) {
		return api.Invite404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if problem, ok := validationFailed(err); ok {
		return api.Invite422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: problem}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.Invite204Response{}, nil
}

// CancelInvitation cancels the invitation of a roster member to the race.
func (controller *ParticipationController) CancelInvitation(ctx context.Context, request api.CancelInvitationRequestObject) (api.CancelInvitationResponseObject, error) {
	command := controller.commandFactory.Invitation(shared.RaceID(request.RaceId), shared.MemberID(request.RosterMemberId))
	err := controller.useCases.CancelInvitation.Execute(ctx, command)
	if errors.Is(err, shared.ErrNotFound) {
		return api.CancelInvitation404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.CancelInvitation204Response{}, nil
}

// AnswerAttendance saves the caller's answer for the race. A caller who is not invited, or
// who already answered "not attending", is 404. An unknown answer is 400, because the
// endpoint has no 422 (app/backend/api/openapi.yaml).
func (controller *ParticipationController) AnswerAttendance(ctx context.Context, request api.AnswerAttendanceRequestObject) (api.AnswerAttendanceResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.AnswerAttendance401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	if request.Body == nil {
		return api.AnswerAttendance400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	command := controller.commandFactory.Attendance(shared.RaceID(request.RaceId), string(request.Body.Status))
	answered, err := controller.useCases.AnswerAttendance.Execute(ctx, caller, command)
	if errors.Is(err, shared.ErrNotFound) {
		return api.AnswerAttendance404ApplicationProblemPlusJSONResponse(notFound()), nil
	}
	var validationErr *shared.ValidationError
	if errors.As(err, &validationErr) {
		return api.AnswerAttendance400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(validationErr.Error())}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.AnswerAttendance200JSONResponse{Status: api.AttendanceStatus(answered.Status())}, nil
}

// ListEntries returns the entries of the race with their members.
func (controller *ParticipationController) ListEntries(ctx context.Context, request api.ListEntriesRequestObject) (api.ListEntriesResponseObject, error) {
	entries, err := controller.useCases.ListEntries.Execute(ctx, shared.RaceID(request.RaceId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.ListEntries404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	response := api.ListEntries200JSONResponse{Entries: make([]api.Entry, 0, len(entries))}
	for _, entry := range entries {
		response.Entries = append(response.Entries, toAPIEntry(entry))
	}
	return response, nil
}

// CreateEntry creates an entry in the race and returns it. A race that already has an
// entry is 409. A blank name is 400, because the endpoint has no 422
// (app/backend/api/openapi.yaml).
func (controller *ParticipationController) CreateEntry(ctx context.Context, request api.CreateEntryRequestObject) (api.CreateEntryResponseObject, error) {
	if request.Body == nil {
		return api.CreateEntry400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	command := controller.commandFactory.CreateEntry(shared.RaceID(request.RaceId), request.Body.Name)
	created, err := controller.useCases.CreateEntry.Execute(ctx, command)
	if errors.Is(err, shared.ErrNotFound) {
		return api.CreateEntry404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if problem, ok := stateConflict(err); ok {
		return api.CreateEntry409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: problem}, nil
	}
	var validationErr *shared.ValidationError
	if errors.As(err, &validationErr) {
		return api.CreateEntry400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(validationErr.Error())}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.CreateEntry201JSONResponse(toAPIEntry(created)), nil
}

// AddEntryMember adds a roster member to the entry. A member whose answer is not
// "attending" is 422.
func (controller *ParticipationController) AddEntryMember(ctx context.Context, request api.AddEntryMemberRequestObject) (api.AddEntryMemberResponseObject, error) {
	command := controller.commandFactory.EntryMember(shared.RaceID(request.RaceId), shared.EntryID(request.EntryId), shared.MemberID(request.RosterMemberId))
	err := controller.useCases.AddEntryMember.Execute(ctx, command)
	if errors.Is(err, shared.ErrNotFound) {
		return api.AddEntryMember404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if problem, ok := validationFailed(err); ok {
		return api.AddEntryMember422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: problem}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.AddEntryMember204Response{}, nil
}

// RemoveEntryMember removes a roster member from the entry.
func (controller *ParticipationController) RemoveEntryMember(ctx context.Context, request api.RemoveEntryMemberRequestObject) (api.RemoveEntryMemberResponseObject, error) {
	command := controller.commandFactory.EntryMember(shared.RaceID(request.RaceId), shared.EntryID(request.EntryId), shared.MemberID(request.RosterMemberId))
	err := controller.useCases.RemoveEntryMember.Execute(ctx, command)
	if errors.Is(err, shared.ErrNotFound) {
		return api.RemoveEntryMember404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.RemoveEntryMember204Response{}, nil
}

func toAPIEntry(entry participations.EntryDTO) api.Entry {
	members := entry.Members()
	body := api.Entry{Id: string(entry.ID()), Name: entry.Name(), Members: make([]api.RosterMemberRef, 0, len(members))}
	for _, member := range members {
		body.Members = append(body.Members, api.RosterMemberRef{RosterMemberId: string(member.MemberID()), Name: member.Name()})
	}
	return body
}
