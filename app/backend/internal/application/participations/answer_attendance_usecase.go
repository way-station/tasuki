package participations

import (
	"context"
	"errors"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// AnswerAttendanceUseCase saves the caller's answer for a race they are invited to
// (PUT /races/{race_id}/attendance). Members may call it; the middleware checks the role.
type AnswerAttendanceUseCase struct {
	unitOfWork  transaction.UnitOfWork
	invitations participation.InvitationRepository
	attendances participation.AttendanceRepository
	entries     participation.EntryRepository
	runners     RunnerRemover
	factory     participation.AttendanceFactory
	dtoFactory  ParticipationDTOFactory
}

// NewAnswerAttendanceUseCase returns an AnswerAttendanceUseCase.
func NewAnswerAttendanceUseCase(unitOfWork transaction.UnitOfWork, invitations participation.InvitationRepository, attendances participation.AttendanceRepository, entries participation.EntryRepository, runners RunnerRemover, factory participation.AttendanceFactory, dtoFactory ParticipationDTOFactory) *AnswerAttendanceUseCase {
	return &AnswerAttendanceUseCase{unitOfWork: unitOfWork, invitations: invitations, attendances: attendances, entries: entries, runners: runners, factory: factory, dtoFactory: dtoFactory}
}

// Execute saves the answer. Changing away from "attending" takes the caller out of the
// entry and off the running orders in the same unit of work. A caller who is not invited,
// such as an organizer without a link, and a caller who already answered "not attending"
// are shared.ErrNotFound (docs/design/api.md).
func (useCase *AnswerAttendanceUseCase) Execute(ctx context.Context, caller auth.MeDTO, command AttendanceCommand) (AttendanceDTO, error) {
	memberID, ok := rosterMemberOf(caller)
	if !ok {
		return AttendanceDTO{}, fmt.Errorf("invitation to race %s: %w", command.raceID, shared.ErrNotFound)
	}
	var answered *participation.Attendance
	err := useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		invitation, err := useCase.invitations.Find(ctx, command.raceID, memberID)
		if err != nil {
			return err
		}
		attendance, leftAttending, err := useCase.answer(ctx, invitation, command.status)
		if err != nil {
			return err
		}
		if err := useCase.attendances.Save(ctx, attendance); err != nil {
			return err
		}
		answered = attendance
		if !leftAttending {
			return nil
		}
		return leaveEntries(ctx, useCase.entries, useCase.runners, command.raceID, memberID)
	})
	if errors.Is(err, participation.ErrAnswerFinal) {
		return AttendanceDTO{}, fmt.Errorf("race %s: %w: %w", command.raceID, shared.ErrNotFound, err)
	}
	if err != nil {
		return AttendanceDTO{}, err
	}
	return useCase.dtoFactory.CreateAttendance(answered.Status()), nil
}

// answer returns the first answer, or the changed answer and whether it changed away from
// "attending".
func (useCase *AnswerAttendanceUseCase) answer(ctx context.Context, invitation *participation.Invitation, status participation.AttendanceStatus) (*participation.Attendance, bool, error) {
	attendance, err := useCase.attendances.Find(ctx, invitation.RaceID(), invitation.MemberID())
	if errors.Is(err, shared.ErrNotFound) {
		first, err := useCase.factory.New(invitation, status)
		return first, false, err
	}
	if err != nil {
		return nil, false, err
	}
	leftAttending, err := attendance.Change(status)
	return attendance, leftAttending, err
}
