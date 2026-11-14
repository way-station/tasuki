package presentation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/application/participations"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

const participationTeamID shared.TeamID = "team-1"

// The session token of each caller in these tests is their user ID. The member acts as
// roster member m1.
const (
	organizerToken = "organizer"
	memberToken    = "member"
	waitingToken   = "waiting"
)

type tokenAuthenticator struct{}

func (tokenAuthenticator) Execute(_ context.Context, token string) (shared.UserID, error) {
	if token == "" {
		return "", auth.ErrUnauthenticated
	}
	return shared.UserID(token), nil
}

type tokenCallers struct{}

func (tokenCallers) Execute(_ context.Context, userID shared.UserID) (auth.MeDTO, error) {
	factory := auth.MeDTOFactory{}
	switch userID {
	case organizerToken:
		return factory.Create(userID, "Organizer", true, identity.CallerNone, nil), nil
	case memberToken:
		return factory.Create(userID, "Member", false, identity.CallerApproved, factory.CreateRosterMember("m1", "Taro")), nil
	}
	return factory.Create(userID, "Waiting", false, identity.CallerPending, nil), nil
}

func sessionOf(token string) *http.Cookie {
	if token == "" {
		return nil
	}
	return &http.Cookie{Name: middleware.SessionCookieName, Value: token, Path: "/api", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode}
}

func problemCode(t *testing.T, body []byte) string {
	t.Helper()
	var problem struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(body, &problem); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return problem.Code
}

type memberKey struct {
	raceID   shared.RaceID
	memberID shared.MemberID
}

// participationStore holds race r1 of the team, the roster, invitations, answers, and
// entries in memory.
type participationStore struct {
	members     []*roster.Member
	invitations map[memberKey]bool
	answers     map[memberKey]participation.AttendanceStatus
	entries     []*participation.Entry
}

// newParticipationStore returns the roster Taro (m1, invited and attending), Jiro (m2), and
// Saburo (m3, who has left), and no entry.
func newParticipationStore() *participationStore {
	leftAt := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	factory := roster.NewMemberFactory(nil)
	return &participationStore{
		members: []*roster.Member{
			factory.Reconstruct("m1", participationTeamID, "Taro", nil),
			factory.Reconstruct("m2", participationTeamID, "Jiro", nil),
			factory.Reconstruct("m3", participationTeamID, "Saburo", &leftAt),
		},
		invitations: map[memberKey]bool{{"r1", "m1"}: true},
		answers:     map[memberKey]participation.AttendanceStatus{{"r1", "m1"}: participation.Attending},
	}
}

func (store *participationStore) withEntry() *participationStore {
	store.entries = append(store.entries, participation.NewEntryFactory(nil).Reconstruct("e1", "r1", "Team A", nil))
	return store
}

func (store *participationStore) memberName(id shared.MemberID) string {
	for _, member := range store.members {
		if member.ID() == id {
			return member.Name()
		}
	}
	return ""
}

func notFoundError(what string, id any) error {
	return fmt.Errorf("%s %v: %w", what, id, shared.ErrNotFound)
}

type storeRaces struct{ store *participationStore }

func (storeRaces) FindByID(_ context.Context, id shared.RaceID) (*race.Race, error) {
	if id != "r1" {
		return nil, notFoundError("race", id)
	}
	return race.NewFactory(nil).Reconstruct("r1", participationTeamID, "Ekiden", race.Description{}, 0), nil
}

func (storeRaces) Save(context.Context, *race.Race) error { return nil }

type storeMembers struct{ store *participationStore }

func (members storeMembers) FindByID(_ context.Context, id shared.MemberID) (*roster.Member, error) {
	for _, member := range members.store.members {
		if member.ID() == id {
			return member, nil
		}
	}
	return nil, notFoundError("roster member", id)
}

func (storeMembers) Save(context.Context, *roster.Member) error { return nil }

type storeInvitations struct{ store *participationStore }

func (invitations storeInvitations) Find(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) (*participation.Invitation, error) {
	if !invitations.store.invitations[memberKey{raceID, memberID}] {
		return nil, notFoundError("invitation of", memberID)
	}
	return participation.InvitationFactory{}.Reconstruct(raceID, memberID), nil
}

func (storeInvitations) ListByMember(context.Context, shared.MemberID) ([]*participation.Invitation, error) {
	return nil, nil
}

func (invitations storeInvitations) Save(_ context.Context, invitation *participation.Invitation) error {
	invitations.store.invitations[memberKey{invitation.RaceID(), invitation.MemberID()}] = true
	return nil
}

func (invitations storeInvitations) Delete(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) error {
	delete(invitations.store.invitations, memberKey{raceID, memberID})
	delete(invitations.store.answers, memberKey{raceID, memberID})
	return nil
}

type storeAttendances struct{ store *participationStore }

func (attendances storeAttendances) Find(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) (*participation.Attendance, error) {
	status, ok := attendances.store.answers[memberKey{raceID, memberID}]
	if !ok {
		return nil, notFoundError("attendance of", memberID)
	}
	return participation.AttendanceFactory{}.Reconstruct(raceID, memberID, status), nil
}

func (attendances storeAttendances) Save(_ context.Context, attendance *participation.Attendance) error {
	attendances.store.answers[memberKey{attendance.RaceID(), attendance.MemberID()}] = attendance.Status()
	return nil
}

type storeEntries struct{ store *participationStore }

func (entries storeEntries) FindByID(_ context.Context, id shared.EntryID) (*participation.Entry, error) {
	for _, entry := range entries.store.entries {
		if entry.ID() == id {
			return participation.NewEntryFactory(nil).Reconstruct(entry.ID(), entry.RaceID(), entry.Name(), entry.Members()), nil
		}
	}
	return nil, notFoundError("entry", id)
}

func (entries storeEntries) ListByRace(ctx context.Context, raceID shared.RaceID) ([]*participation.Entry, error) {
	var listed []*participation.Entry
	for _, entry := range entries.store.entries {
		if entry.RaceID() == raceID {
			found, err := entries.FindByID(ctx, entry.ID())
			if err != nil {
				return nil, err
			}
			listed = append(listed, found)
		}
	}
	return listed, nil
}

func (entries storeEntries) Save(_ context.Context, saved *participation.Entry) error {
	for index, entry := range entries.store.entries {
		if entry.ID() == saved.ID() {
			entries.store.entries[index] = saved
			return nil
		}
	}
	entries.store.entries = append(entries.store.entries, saved)
	return nil
}

type storeQueries struct{ store *participationStore }

func (queries storeQueries) ListInvitations(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]participations.InvitationRecord, error) {
	if _, err := (storeRaces{}).FindByID(ctx, raceID); err != nil || teamID != participationTeamID {
		return nil, notFoundError("race", raceID)
	}
	var records []participations.InvitationRecord
	for _, member := range queries.store.members {
		if member.Left() {
			continue
		}
		key := memberKey{raceID, member.ID()}
		record := participations.InvitationRecord{MemberID: member.ID(), Name: member.Name(), Invited: queries.store.invitations[key]}
		if status, ok := queries.store.answers[key]; ok {
			record.Attendance = &status
		}
		records = append(records, record)
	}
	return records, nil
}

func (queries storeQueries) ListEntries(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]participations.EntryRecord, error) {
	if _, err := (storeRaces{}).FindByID(ctx, raceID); err != nil || teamID != participationTeamID {
		return nil, notFoundError("race", raceID)
	}
	records := []participations.EntryRecord{}
	for _, entry := range queries.store.entries {
		record := participations.EntryRecord{ID: entry.ID(), Name: entry.Name()}
		for _, memberID := range entry.Members() {
			record.Members = append(record.Members, participations.EntryMemberRecord{MemberID: memberID, Name: queries.store.memberName(memberID)})
		}
		records = append(records, record)
	}
	return records, nil
}

type passUnitOfWork struct{}

func (passUnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	return work(ctx)
}

type passRunnerRemover struct{}

func (passRunnerRemover) RemoveRunner(context.Context, shared.RaceID, shared.MemberID) error {
	return nil
}

func newParticipationRouter(store *participationStore) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	races, members, invitations := storeRaces{store}, storeMembers{store}, storeInvitations{store}
	attendances, entries, queries := storeAttendances{store}, storeEntries{store}, storeQueries{store}
	dtoFactory := participations.ParticipationDTOFactory{}
	participationController := controller.NewParticipationController(controller.ParticipationUseCases{
		ListInvitations:   participations.NewListInvitationsUseCase(queries, participation.AttendancePolicy{}, dtoFactory, participationTeamID),
		Invite:            participations.NewInviteUseCase(races, members, invitations, participation.InvitationFactory{}, participationTeamID),
		CancelInvitation:  participations.NewCancelInvitationUseCase(passUnitOfWork{}, races, invitations, entries, passRunnerRemover{}, participationTeamID),
		AnswerAttendance:  participations.NewAnswerAttendanceUseCase(passUnitOfWork{}, invitations, attendances, entries, passRunnerRemover{}, participation.AttendanceFactory{}, dtoFactory),
		ListEntries:       participations.NewListEntriesUseCase(queries, dtoFactory, participationTeamID),
		CreateEntry:       participations.NewCreateEntryUseCase(races, entries, participation.NewEntryFactory(&sequentialIDs{}), dtoFactory, participationTeamID),
		AddEntryMember:    participations.NewAddEntryMemberUseCase(races, entries, attendances, participationTeamID),
		RemoveEntryMember: participations.NewRemoveEntryMemberUseCase(passUnitOfWork{}, races, entries, passRunnerRemover{}, participationTeamID),
	}, participations.ParticipationCommandFactory{})
	check := health.NewCheckUseCase(fakeHealth{}, fakeHealth{}, health.StatusDTOFactory{}, logger)
	return NewRouter(Controllers{Health: controller.NewHealthController(check), Participation: participationController}, Options{
		Authenticate:   middleware.Authenticate(tokenAuthenticator{}, logger),
		ResolveCaller:  middleware.ResolveCaller(tokenCallers{}, logger),
		FrontendOrigin: testOrigin,
		Logger:         logger,
	})
}

func TestParticipationRoutesCallers(t *testing.T) {
	tests := []struct {
		method, path, body string
		want               map[string]int // by session token; "" is logged out
	}{
		{method: http.MethodGet, path: "/api/v1/races/r1/invitations", want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 200}},
		{method: http.MethodPut, path: "/api/v1/races/r1/invitations/m2", want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 204}},
		{method: http.MethodDelete, path: "/api/v1/races/r1/invitations/m1", want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 204}},
		{method: http.MethodPut, path: "/api/v1/races/r1/attendance", body: `{"status":"undecided"}`, want: map[string]int{"": 401, waitingToken: 403, memberToken: 200, organizerToken: 404}},
		{method: http.MethodGet, path: "/api/v1/races/r1/entries", want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 200}},
		{method: http.MethodPost, path: "/api/v1/races/r1/entries", body: `{"name":"Team B"}`, want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 409}},
		{method: http.MethodPut, path: "/api/v1/races/r1/entries/e1/members/m1", want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 204}},
		{method: http.MethodDelete, path: "/api/v1/races/r1/entries/e1/members/m1", want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 204}},
	}
	for _, testCase := range tests {
		for _, token := range []string{"", waitingToken, memberToken, organizerToken} {
			t.Run(fmt.Sprintf("%s %s as %q", testCase.method, testCase.path, token), func(t *testing.T) {
				handler := newParticipationRouter(newParticipationStore().withEntry())
				rec := serve(handler, testCase.method, testCase.path, testCase.body, sessionOf(token))
				if rec.Code != testCase.want[token] {
					t.Errorf("status = %d, want %d (body %s)", rec.Code, testCase.want[token], rec.Body)
				}
			})
		}
	}
}

func TestParticipationRoutesBodies(t *testing.T) {
	t.Run("lists the members who have not left with their invitation", func(t *testing.T) {
		rec := serve(newParticipationRouter(newParticipationStore()), http.MethodGet, "/api/v1/races/r1/invitations", "", sessionOf(organizerToken))
		want := `{"members":[{"attendance":"attending","invited":true,"name":"Taro","roster_member_id":"m1"},` +
			`{"attendance":null,"invited":false,"name":"Jiro","roster_member_id":"m2"}]}`
		if rec.Code != http.StatusOK || rec.Body.String() != want+"\n" {
			t.Errorf("status %d, body %s\nwant %s", rec.Code, rec.Body, want)
		}
	})

	t.Run("returns the answer", func(t *testing.T) {
		rec := serve(newParticipationRouter(newParticipationStore()), http.MethodPut, "/api/v1/races/r1/attendance", `{"status":"undecided"}`, sessionOf(memberToken))
		if want := `{"status":"undecided"}`; rec.Code != http.StatusOK || rec.Body.String() != want+"\n" {
			t.Errorf("status %d, body %s, want %s", rec.Code, rec.Body, want)
		}
	})

	t.Run("returns the created entry, then lists it with its members", func(t *testing.T) {
		handler := newParticipationRouter(newParticipationStore())
		rec := serve(handler, http.MethodPost, "/api/v1/races/r1/entries", `{"name":"Team A"}`, sessionOf(organizerToken))
		if want := `{"id":"id1","members":[],"name":"Team A"}`; rec.Code != http.StatusCreated || rec.Body.String() != want+"\n" {
			t.Fatalf("POST status %d, body %s, want %s", rec.Code, rec.Body, want)
		}
		if rec := serve(handler, http.MethodPut, "/api/v1/races/r1/entries/id1/members/m1", "", sessionOf(organizerToken)); rec.Code != http.StatusNoContent {
			t.Fatalf("PUT member status %d, body %s", rec.Code, rec.Body)
		}
		rec = serve(handler, http.MethodGet, "/api/v1/races/r1/entries", "", sessionOf(organizerToken))
		if want := `{"entries":[{"id":"id1","members":[{"name":"Taro","roster_member_id":"m1"}],"name":"Team A"}]}`; rec.Code != http.StatusOK || rec.Body.String() != want+"\n" {
			t.Errorf("GET status %d, body %s, want %s", rec.Code, rec.Body, want)
		}
	})
}

func TestParticipationRoutesErrors(t *testing.T) {
	tests := []struct {
		name, method, path, body, token string
		store                           *participationStore
		wantStatus                      int
		wantCode                        string
	}{
		{name: "a race of no team", method: http.MethodGet, path: "/api/v1/races/r9/invitations", token: organizerToken, store: newParticipationStore(), wantStatus: 404, wantCode: "not_found"},
		{name: "inviting a member who has left", method: http.MethodPut, path: "/api/v1/races/r1/invitations/m3", token: organizerToken, store: newParticipationStore(), wantStatus: 422, wantCode: "validation_failed"},
		{name: "inviting a member who does not exist", method: http.MethodPut, path: "/api/v1/races/r1/invitations/m9", token: organizerToken, store: newParticipationStore(), wantStatus: 404, wantCode: "not_found"},
		{name: "an unknown answer", method: http.MethodPut, path: "/api/v1/races/r1/attendance", body: `{"status":"maybe"}`, token: memberToken, store: newParticipationStore(), wantStatus: 400, wantCode: "invalid_request"},
		{name: "an answer without a body", method: http.MethodPut, path: "/api/v1/races/r1/attendance", token: memberToken, store: newParticipationStore(), wantStatus: 400, wantCode: "invalid_request"},
		{name: "an answer after not attending", method: http.MethodPut, path: "/api/v1/races/r1/attendance", body: `{"status":"attending"}`, token: memberToken, store: notAttendingStore(), wantStatus: 404, wantCode: "not_found"},
		{name: "an entry with a blank name", method: http.MethodPost, path: "/api/v1/races/r1/entries", body: `{"name":" "}`, token: organizerToken, store: newParticipationStore(), wantStatus: 400, wantCode: "invalid_request"},
		{name: "a second entry", method: http.MethodPost, path: "/api/v1/races/r1/entries", body: `{"name":"Team B"}`, token: organizerToken, store: newParticipationStore().withEntry(), wantStatus: 409, wantCode: "state_conflict"},
		{name: "adding a member who is not attending", method: http.MethodPut, path: "/api/v1/races/r1/entries/e1/members/m2", token: organizerToken, store: newParticipationStore().withEntry(), wantStatus: 422, wantCode: "validation_failed"},
		{name: "an entry that does not exist", method: http.MethodDelete, path: "/api/v1/races/r1/entries/e9/members/m1", token: organizerToken, store: newParticipationStore().withEntry(), wantStatus: 404, wantCode: "not_found"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := serve(newParticipationRouter(testCase.store), testCase.method, testCase.path, testCase.body, sessionOf(testCase.token))
			if rec.Code != testCase.wantStatus || problemCode(t, rec.Body.Bytes()) != testCase.wantCode {
				t.Errorf("status %d, body %s, want %d %s", rec.Code, rec.Body, testCase.wantStatus, testCase.wantCode)
			}
		})
	}
}

func notAttendingStore() *participationStore {
	store := newParticipationStore()
	store.answers[memberKey{"r1", "m1"}] = participation.NotAttending
	return store
}
