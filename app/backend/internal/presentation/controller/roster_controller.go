package controller

import (
	"context"
	"errors"

	"github.com/way-station/tasuki/app/backend/internal/application/rosters"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
)

// RosterController serves the roster and the approval of member links (docs/design/api.md).
// The routes let only organizers through; the use cases keep to the team.
type RosterController struct {
	list             *rosters.ListRosterMembersUseCase
	add              *rosters.AddRosterMemberUseCase
	get              *rosters.GetRosterMemberUseCase
	setLeft          *rosters.SetLeftUseCase
	unlink           *rosters.UnlinkUseCase
	createInviteLink *rosters.CreateInviteLinkUseCase
	listPending      *rosters.ListPendingLinksUseCase
	decide           *rosters.DecideLinkUseCase
	commandFactory   rosters.RosterCommandFactory
	frontendOrigin   string
}

// NewRosterController returns a RosterController. The invite link URLs start with
// frontendOrigin (FRONTEND_ORIGIN).
func NewRosterController(list *rosters.ListRosterMembersUseCase, add *rosters.AddRosterMemberUseCase, get *rosters.GetRosterMemberUseCase, setLeft *rosters.SetLeftUseCase, unlink *rosters.UnlinkUseCase, createInviteLink *rosters.CreateInviteLinkUseCase, listPending *rosters.ListPendingLinksUseCase, decide *rosters.DecideLinkUseCase, commandFactory rosters.RosterCommandFactory, frontendOrigin string) *RosterController {
	return &RosterController{list: list, add: add, get: get, setLeft: setLeft, unlink: unlink, createInviteLink: createInviteLink, listPending: listPending, decide: decide, commandFactory: commandFactory, frontendOrigin: frontendOrigin}
}

// ListRosterMembers returns the roster by name.
func (controller *RosterController) ListRosterMembers(ctx context.Context, _ api.ListRosterMembersRequestObject) (api.ListRosterMembersResponseObject, error) {
	members, err := controller.list.Execute(ctx)
	if err != nil {
		return nil, err
	}
	apiMembers := make([]api.RosterMember, 0, len(members))
	for _, member := range members {
		apiMembers = append(apiMembers, toAPIRosterMember(member))
	}
	return api.ListRosterMembers200JSONResponse{RosterMembers: apiMembers}, nil
}

// CreateRosterMember adds a roster member. An empty name is 400 (docs/design/api.md).
func (controller *RosterController) CreateRosterMember(ctx context.Context, request api.CreateRosterMemberRequestObject) (api.CreateRosterMemberResponseObject, error) {
	if request.Body == nil {
		return api.CreateRosterMember400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	member, err := controller.add.Execute(ctx, controller.commandFactory.AddRosterMember(request.Body.Name))
	var validationErr *shared.ValidationError
	if errors.As(err, &validationErr) {
		return api.CreateRosterMember400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(validationErr.Error())}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.CreateRosterMember201JSONResponse(toAPIRosterMember(member)), nil
}

// GetRosterMember returns the member's profile with their laps by race. A member of another
// team is 404.
func (controller *RosterController) GetRosterMember(ctx context.Context, request api.GetRosterMemberRequestObject) (api.GetRosterMemberResponseObject, error) {
	member, err := controller.get.Execute(ctx, shared.MemberID(request.RosterMemberId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.GetRosterMember404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.GetRosterMember200JSONResponse(toAPIRosterMemberDetail(member)), nil
}

// UpdateRosterMember marks the member as left, or undoes it, and returns the member. A member
// of another team is 404.
func (controller *RosterController) UpdateRosterMember(ctx context.Context, request api.UpdateRosterMemberRequestObject) (api.UpdateRosterMemberResponseObject, error) {
	if request.Body == nil {
		return api.UpdateRosterMember400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	memberID := shared.MemberID(request.RosterMemberId)
	err := controller.setLeft.Execute(ctx, controller.commandFactory.SetLeft(memberID, request.Body.Left))
	if errors.Is(err, shared.ErrNotFound) {
		return api.UpdateRosterMember404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	member, err := controller.list.Find(ctx, memberID)
	if err != nil {
		return nil, err
	}
	return api.UpdateRosterMember200JSONResponse(toAPIRosterMember(member)), nil
}

// UnlinkRosterMember removes the member's approved link. A member without one is 409, and a
// member of another team is 404.
func (controller *RosterController) UnlinkRosterMember(ctx context.Context, request api.UnlinkRosterMemberRequestObject) (api.UnlinkRosterMemberResponseObject, error) {
	err := controller.unlink.Execute(ctx, shared.MemberID(request.RosterMemberId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.UnlinkRosterMember404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if conflict, ok := stateConflict(err); ok {
		return api.UnlinkRosterMember409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.UnlinkRosterMember204Response{}, nil
}

// CreateInviteLink replaces the member's invite link with a new one and returns its URL,
// FRONTEND_ORIGIN followed by /invite/<token>. A member who has left, or has an approved or
// pending link, is 409, and a member of another team is 404.
func (controller *RosterController) CreateInviteLink(ctx context.Context, request api.CreateInviteLinkRequestObject) (api.CreateInviteLinkResponseObject, error) {
	link, err := controller.createInviteLink.Execute(ctx, shared.MemberID(request.RosterMemberId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.CreateInviteLink404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if conflict, ok := stateConflict(err); ok {
		return api.CreateInviteLink409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.CreateInviteLink201JSONResponse{Url: controller.frontendOrigin + "/invite/" + link.Token(), ExpiresAt: link.ExpiresAt()}, nil
}

// ListPendingMemberLinks returns the links that wait for approval.
func (controller *RosterController) ListPendingMemberLinks(ctx context.Context, _ api.ListPendingMemberLinksRequestObject) (api.ListPendingMemberLinksResponseObject, error) {
	links, err := controller.listPending.Execute(ctx)
	if err != nil {
		return nil, err
	}
	apiLinks := make([]api.MemberLink, 0, len(links))
	for _, link := range links {
		apiLinks = append(apiLinks, api.MemberLink{
			Id:              string(link.LinkID()),
			LineDisplayName: link.LineDisplayName(),
			RosterMember:    api.RosterMemberBasic{Id: string(link.MemberID()), Name: link.MemberName()},
			CreatedAt:       link.CreatedAt(),
		})
	}
	return api.ListPendingMemberLinks200JSONResponse{Links: apiLinks}, nil
}

// ApproveMemberLink approves a pending link. A link that is not pending is 409, and a link
// of a member of another team is 404.
func (controller *RosterController) ApproveMemberLink(ctx context.Context, request api.ApproveMemberLinkRequestObject) (api.ApproveMemberLinkResponseObject, error) {
	err := controller.decide.Approve(ctx, shared.MemberLinkID(request.LinkId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.ApproveMemberLink404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if conflict, ok := stateConflict(err); ok {
		return api.ApproveMemberLink409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.ApproveMemberLink204Response{}, nil
}

// RejectMemberLink rejects a pending link. A link that is not pending is 409, and a link of
// a member of another team is 404.
func (controller *RosterController) RejectMemberLink(ctx context.Context, request api.RejectMemberLinkRequestObject) (api.RejectMemberLinkResponseObject, error) {
	err := controller.decide.Reject(ctx, shared.MemberLinkID(request.LinkId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.RejectMemberLink404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if conflict, ok := stateConflict(err); ok {
		return api.RejectMemberLink409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.RejectMemberLink204Response{}, nil
}

func toAPIRosterMember(dto rosters.RosterMemberDTO) api.RosterMember {
	member := api.RosterMember{Id: string(dto.MemberID()), Name: dto.Name(), Left: dto.Left(), LinkStatus: toAPILinkStatus(dto.LinkStatus())}
	if expiresAt, ok := dto.InviteLinkExpiresAt(); ok {
		member.InviteLinkExpiresAt = &expiresAt
	}
	return member
}

func toAPILinkStatus(status roster.LinkStatus) api.RosterMemberLinkStatus {
	switch status {
	case roster.LinkPending:
		return api.RosterMemberLinkStatusPending
	case roster.LinkApproved:
		return api.RosterMemberLinkStatusApproved
	default:
		return api.RosterMemberLinkStatusNone
	}
}

func toAPIRosterMemberDetail(dto rosters.RosterMemberDetailDTO) api.RosterMemberDetail {
	lapsByRace := make([]api.RaceLapsOfMember, 0, len(dto.LapsByRace()))
	for _, race := range dto.LapsByRace() {
		raceLaps := make([]api.Lap, 0, len(race.Laps()))
		for _, raceLap := range race.Laps() {
			raceLaps = append(raceLaps, api.Lap{Id: string(raceLap.ID()), Time: raceLap.Time().String()})
		}
		lapsByRace = append(lapsByRace, api.RaceLapsOfMember{RaceId: string(race.RaceID()), RaceName: race.RaceName(), Laps: raceLaps})
	}
	return api.RosterMemberDetail{Id: string(dto.MemberID()), Name: dto.Name(), Left: dto.Left(), AverageTime: dto.Average().String(), LapCount: dto.LapCount(), LapsByRace: lapsByRace}
}
