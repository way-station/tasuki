package presentation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/application/laps"
	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

const lapTeamID shared.TeamID = "team-1"

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

type passUnitOfWork struct{}

func (passUnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	return work(ctx)
}

// lapStore holds race r1 of the team, whose confirmed running order has Taro (m1) and Jiro
// (m2), the lap records, and the members' averages. The outbox runs the steps right away,
// as the request does after the commit.
type lapStore struct {
	runners []shared.MemberID
	records map[shared.MemberID]*lap.Record
	members map[shared.MemberID]*roster.Member
}

func newLapStore(runners ...shared.MemberID) *lapStore {
	factory := roster.NewMemberFactory(&sequentialIDs{})
	return &lapStore{runners: runners, records: map[shared.MemberID]*lap.Record{}, members: map[shared.MemberID]*roster.Member{
		"m1": factory.Reconstruct("m1", lapTeamID, "Taro", nil, roster.DefaultAverage(), 0),
		"m2": factory.Reconstruct("m2", lapTeamID, "Jiro", nil, roster.DefaultAverage(), 0),
	}}
}

// withLaps stores the member's laps for r1, with the IDs l1, l2, and so on.
func (store *lapStore) withLaps(memberID shared.MemberID, times ...string) *lapStore {
	record := lap.RecordFactory{}.New("r1", memberID)
	var all []shared.Pace
	for i, text := range times {
		lapTime, _ := shared.ParsePace(text)
		record.Add(lap.NewFactory(nil).Reconstruct(shared.LapID(fmt.Sprintf("l%d", i+1)), lapTime, time.Time{}))
		all = append(all, lapTime)
	}
	store.records[memberID] = record
	roster.AverageService{}.Recalculate(store.members[memberID], all)
	return store
}

type storeLapRecords struct{ store *lapStore }

func (records storeLapRecords) Find(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) (*lap.Record, error) {
	record, ok := records.store.records[memberID]
	if raceID != "r1" || !ok {
		return nil, fmt.Errorf("lap record of %s in %s: %w", memberID, raceID, shared.ErrNotFound)
	}
	return lap.RecordFactory{}.Reconstruct(record.RaceID(), record.MemberID(), record.Laps()), nil
}

func (records storeLapRecords) ListByMember(_ context.Context, memberID shared.MemberID) ([]*lap.Record, error) {
	if record, ok := records.store.records[memberID]; ok {
		return []*lap.Record{record}, nil
	}
	return nil, nil
}

func (records storeLapRecords) Save(_ context.Context, record *lap.Record) error {
	records.store.records[record.MemberID()] = record
	return nil
}

func (records storeLapRecords) Delete(_ context.Context, _ shared.RaceID, memberID shared.MemberID) error {
	delete(records.store.records, memberID)
	return nil
}

type storeLapMembers struct{ store *lapStore }

func (members storeLapMembers) FindByID(_ context.Context, id shared.MemberID) (*roster.Member, error) {
	member, ok := members.store.members[id]
	if !ok {
		return nil, fmt.Errorf("member %s: %w", id, shared.ErrNotFound)
	}
	return member, nil
}

func (members storeLapMembers) Save(_ context.Context, member *roster.Member) error {
	members.store.members[member.ID()] = member
	return nil
}

// runningOutbox runs the steps of an entry when it is added.
type runningOutbox struct{ runner *laps.LapStepRunner }

func (store runningOutbox) Add(ctx context.Context, entry outbox.Entry) error {
	for _, step := range entry.Steps() {
		var err error
		switch typed := step.(type) {
		case laps.AddLapsStep:
			err = store.runner.AddLaps(ctx, typed)
		case laps.EditLapStep:
			err = store.runner.EditLap(ctx, typed)
		case laps.DeleteLapStep:
			err = store.runner.DeleteLap(ctx, typed)
		case laps.RecalculateAverageStep:
			err = store.runner.RecalculateAverage(ctx, typed)
		default:
			err = fmt.Errorf("unknown step %q", step.Kind())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type storeLapQueries struct{ store *lapStore }

func (queries storeLapQueries) IsConfirmedRunner(_ context.Context, teamID shared.TeamID, raceID shared.RaceID, memberID shared.MemberID) (bool, error) {
	return teamID == lapTeamID && raceID == "r1" && slices.Contains(queries.store.runners, memberID), nil
}

func (queries storeLapQueries) GetMyLaps(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) (laps.MyLapsRecord, error) {
	member := queries.store.members[memberID]
	return laps.MyLapsRecord{Laps: queries.lapsOf(raceID, memberID), Average: member.Average(), LapCount: member.LapCount()}, nil
}

func (queries storeLapQueries) ListRaceLaps(_ context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]laps.RunnerLapsRecord, error) {
	if teamID != lapTeamID || raceID != "r1" {
		return nil, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	var runners []laps.RunnerLapsRecord
	for _, memberID := range queries.store.runners {
		runners = append(runners, laps.RunnerLapsRecord{MemberID: memberID, Name: queries.store.members[memberID].Name(), Laps: queries.lapsOf(raceID, memberID)})
	}
	return runners, nil
}

func (queries storeLapQueries) GetProfile(_ context.Context, memberID shared.MemberID) (laps.ProfileRecord, error) {
	member, ok := queries.store.members[memberID]
	if !ok {
		return laps.ProfileRecord{}, fmt.Errorf("member %s: %w", memberID, shared.ErrNotFound)
	}
	return laps.ProfileRecord{Name: member.Name(), Average: member.Average(), LapCount: member.LapCount()}, nil
}

func (queries storeLapQueries) lapsOf(raceID shared.RaceID, memberID shared.MemberID) []laps.LapRecord {
	records := []laps.LapRecord{}
	if record, ok := queries.store.records[memberID]; ok && raceID == "r1" {
		for _, stored := range record.Laps() {
			records = append(records, laps.LapRecord{ID: stored.ID(), Time: stored.Time()})
		}
	}
	return records
}

func newLapRouter(store *lapStore) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	now := func() time.Time { return time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC) }
	records, queries := storeLapRecords{store}, storeLapQueries{store}
	runner := laps.NewLapStepRunner(records, storeLapMembers{store}, lap.NewFactory(nil), lap.RecordFactory{}, roster.AverageService{})
	steps := runningOutbox{runner}
	lapController := controller.NewLapController(
		laps.NewGetMyLapsUseCase(queries, laps.LapDTOFactory{}, lapTeamID),
		laps.NewSubmitLapsUseCase(passUnitOfWork{}, queries, steps, lap.SubmissionPolicy{}, lap.NewFactory(&sequentialIDs{}), outbox.EntryFactory{}, now, lapTeamID),
		laps.NewEditLapUseCase(passUnitOfWork{}, queries, records, steps, lap.SubmissionPolicy{}, outbox.EntryFactory{}, lapTeamID),
		laps.NewDeleteLapUseCase(passUnitOfWork{}, queries, records, steps, lap.SubmissionPolicy{}, outbox.EntryFactory{}, lapTeamID),
		laps.NewListRaceLapsUseCase(queries, lap.VisibilityPolicy{}, laps.LapDTOFactory{}, lapTeamID),
		laps.NewGetProfileUseCase(queries, laps.LapDTOFactory{}),
		laps.LapCommandFactory{},
	)
	check := health.NewCheckUseCase(fakeHealth{}, fakeHealth{}, health.StatusDTOFactory{}, logger)
	return NewRouter(Controllers{Health: controller.NewHealthController(check), Laps: lapController}, Options{
		Authenticate:   middleware.Authenticate(tokenAuthenticator{}, logger),
		ResolveCaller:  middleware.ResolveCaller(tokenCallers{}, logger),
		FrontendOrigin: testOrigin,
		Logger:         logger,
	})
}

func TestLapRoutesCallers(t *testing.T) {
	runners := func() *lapStore { return newLapStore("m1", "m2").withLaps("m1", "3:30") }
	tests := []struct {
		method, path, body string
		want               map[string]int // by session token; "" is logged out
	}{
		// An organizer without a link is not a runner, so the laps are 404 to them (docs/design/api.md).
		{method: http.MethodGet, path: "/api/v1/races/r1/my-laps", want: map[string]int{"": 401, waitingToken: 403, memberToken: 200, organizerToken: 404}},
		{method: http.MethodPost, path: "/api/v1/races/r1/my-laps", body: `{"times":["3:20"]}`, want: map[string]int{"": 401, waitingToken: 403, memberToken: 201, organizerToken: 404}},
		{method: http.MethodPatch, path: "/api/v1/races/r1/my-laps/l1", body: `{"time":"3:20"}`, want: map[string]int{"": 401, waitingToken: 403, memberToken: 200, organizerToken: 404}},
		{method: http.MethodDelete, path: "/api/v1/races/r1/my-laps/l1", want: map[string]int{"": 401, waitingToken: 403, memberToken: 200, organizerToken: 404}},
		{method: http.MethodGet, path: "/api/v1/races/r1/laps", want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 200}},
		{method: http.MethodGet, path: "/api/v1/me/profile", want: map[string]int{"": 401, waitingToken: 403, memberToken: 200, organizerToken: 403}},
	}
	for _, testCase := range tests {
		for _, token := range []string{"", waitingToken, memberToken, organizerToken} {
			t.Run(fmt.Sprintf("%s %s as %q", testCase.method, testCase.path, token), func(t *testing.T) {
				rec := serve(newLapRouter(runners()), testCase.method, testCase.path, testCase.body, sessionOf(token))
				if rec.Code != testCase.want[token] {
					t.Errorf("status = %d, want %d (body %s)", rec.Code, testCase.want[token], rec.Body)
				}
			})
		}
	}
}

func TestLapRoutesBodies(t *testing.T) {
	tests := []struct {
		name, method, path, body, token string
		store                           *lapStore
		wantStatus                      int
		want                            string
	}{
		{name: "returns the submitted laps with the recalculated average", method: http.MethodPost, path: "/api/v1/races/r1/my-laps", body: `{"times":["3:20","3:25"]}`, token: memberToken,
			store: newLapStore("m1").withLaps("m1", "3:30"), wantStatus: 201,
			want: `{"average_time":"3:25","lap_count":3,"laps":[{"id":"l1","time":"3:30"},{"id":"id1","time":"3:20"},{"id":"id2","time":"3:25"}]}`},
		{name: "returns the laps after an edit", method: http.MethodPatch, path: "/api/v1/races/r1/my-laps/l2", body: `{"time":"3:20"}`, token: memberToken,
			store: newLapStore("m1").withLaps("m1", "3:30", "3:40"), wantStatus: 200,
			want: `{"average_time":"3:25","lap_count":2,"laps":[{"id":"l1","time":"3:30"},{"id":"l2","time":"3:20"}]}`},
		{name: "returns the default average after the last lap is deleted", method: http.MethodDelete, path: "/api/v1/races/r1/my-laps/l1", token: memberToken,
			store: newLapStore("m1").withLaps("m1", "3:30"), wantStatus: 200,
			want: `{"average_time":"3:40","lap_count":0,"laps":[]}`},
		{name: "lists every runner for the organizer, with empty laps for runners without any", method: http.MethodGet, path: "/api/v1/races/r1/laps", token: organizerToken,
			store: newLapStore("m2", "m1").withLaps("m1", "3:30"), wantStatus: 200,
			want: `{"runners":[{"laps":[],"name":"Jiro","roster_member_id":"m2"},{"laps":[{"id":"l1","time":"3:30"}],"name":"Taro","roster_member_id":"m1"}]}`},
		{name: "returns the profile", method: http.MethodGet, path: "/api/v1/me/profile", token: memberToken,
			store: newLapStore().withLaps("m1", "3:20", "3:25"), wantStatus: 200,
			want: `{"average_time":"3:23","lap_count":2,"name":"Taro"}`},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := serve(newLapRouter(testCase.store), testCase.method, testCase.path, testCase.body, sessionOf(testCase.token))
			if rec.Code != testCase.wantStatus || rec.Body.String() != testCase.want+"\n" {
				t.Errorf("status %d, body %s\nwant %d %s", rec.Code, rec.Body, testCase.wantStatus, testCase.want)
			}
		})
	}
}

func TestLapRoutesErrors(t *testing.T) {
	tests := []struct {
		name, method, path, body string
		store                    *lapStore
		wantStatus               int
		wantCode                 string
	}{
		{name: "a member who is not a confirmed runner", method: http.MethodGet, path: "/api/v1/races/r1/my-laps", store: newLapStore("m2"), wantStatus: 404, wantCode: "not_found"},
		{name: "laps of another race", method: http.MethodPost, path: "/api/v1/races/r9/my-laps", body: `{"times":["3:20"]}`, store: newLapStore("m1"), wantStatus: 404, wantCode: "not_found"},
		{name: "a time that is not MM:SS", method: http.MethodPost, path: "/api/v1/races/r1/my-laps", body: `{"times":["3:20","320"]}`, store: newLapStore("m1"), wantStatus: 400, wantCode: "invalid_request"},
		{name: "no time", method: http.MethodPost, path: "/api/v1/races/r1/my-laps", body: `{"times":[]}`, store: newLapStore("m1"), wantStatus: 400, wantCode: "invalid_request"},
		{name: "a submission without a body", method: http.MethodPost, path: "/api/v1/races/r1/my-laps", store: newLapStore("m1"), wantStatus: 400, wantCode: "invalid_request"},
		{name: "an edit to a time that is not MM:SS", method: http.MethodPatch, path: "/api/v1/races/r1/my-laps/l1", body: `{"time":"3:2"}`, store: newLapStore("m1").withLaps("m1", "3:30"), wantStatus: 400, wantCode: "invalid_request"},
		{name: "an edit of a lap the member does not have", method: http.MethodPatch, path: "/api/v1/races/r1/my-laps/l9", body: `{"time":"3:20"}`, store: newLapStore("m1").withLaps("m1", "3:30"), wantStatus: 404, wantCode: "not_found"},
		{name: "a deletion of a lap the member does not have", method: http.MethodDelete, path: "/api/v1/races/r1/my-laps/l9", store: newLapStore("m1").withLaps("m1", "3:30"), wantStatus: 404, wantCode: "not_found"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := serve(newLapRouter(testCase.store), testCase.method, testCase.path, testCase.body, sessionOf(memberToken))
			if rec.Code != testCase.wantStatus || problemCode(t, rec.Body.Bytes()) != testCase.wantCode {
				t.Errorf("status %d, body %s, want %d %s", rec.Code, rec.Body, testCase.wantStatus, testCase.wantCode)
			}
		})
	}

	t.Run("a race of another team for the organizer", func(t *testing.T) {
		rec := serve(newLapRouter(newLapStore("m1")), http.MethodGet, "/api/v1/races/r9/laps", "", sessionOf(organizerToken))
		if rec.Code != http.StatusNotFound || problemCode(t, rec.Body.Bytes()) != "not_found" {
			t.Errorf("status %d, body %s, want 404 not_found", rec.Code, rec.Body)
		}
	})
}
