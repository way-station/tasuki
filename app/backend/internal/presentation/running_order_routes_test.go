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
	"github.com/way-station/tasuki/app/backend/internal/application/runningorders"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

const runningOrderTeamID shared.TeamID = "team-1"

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

// runningOrderStore holds the distance race r1 (5 km and 3.2 km) of the team, its entry e1
// of Taro (m1, 3:40/km) and Jiro (m2, 3:30/km), the running order of e1, and the answer of m1.
type runningOrderStore struct {
	answer       race.Answer
	runningOrder *runningorder.RunningOrder
	history      int
}

func newRunningOrderStore(answer race.Answer) *runningOrderStore {
	return &runningOrderStore{answer: answer}
}

// withOrder stores the running order of e1 with a runner per leg (nil for empty).
func (store *runningOrderStore) withOrder(status runningorder.Status, runners ...*shared.MemberID) *runningOrderStore {
	legs := make([]runningorder.Assignment, 0, len(runners))
	for i, runner := range runners {
		legs = append(legs, runningorder.Assignment{Number: i + 1, Runner: runner})
	}
	store.runningOrder = runningorder.NewFactory(nil).Reconstruct("o1", "e1", status, 0, legs)
	return store
}

func memberIDOf(id shared.MemberID) *shared.MemberID { return &id }

var runningOrderRunners = map[shared.MemberID]runningorders.RunnerRecord{
	"m1": {MemberID: "m1", Name: "Taro", AverageSeconds: 220},
	"m2": {MemberID: "m2", Name: "Jiro", AverageSeconds: 210},
}

func runningOrderDescription() race.DescriptionParams {
	raceDate, _ := shared.ParseDate("2026-12-06")
	return race.DescriptionParams{Type: race.Distance, RaceDate: raceDate, Deadline: time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC), LegDistancesKm: []float64{5, 3.2}}
}

type storeRunningOrderRaces struct{}

func (storeRunningOrderRaces) FindByID(_ context.Context, id shared.RaceID) (*race.Race, error) {
	if id != "r1" {
		return nil, fmt.Errorf("race %s: %w", id, shared.ErrNotFound)
	}
	description, err := race.NewDescription(runningOrderDescription())
	if err != nil {
		return nil, err
	}
	return race.NewFactory(nil).Reconstruct("r1", runningOrderTeamID, "Ekiden", description, 0), nil
}

func (storeRunningOrderRaces) Save(context.Context, *race.Race) error { return nil }

type storeRunningOrderEntries struct{}

func (storeRunningOrderEntries) FindByID(_ context.Context, id shared.EntryID) (*participation.Entry, error) {
	if id != "e1" {
		return nil, fmt.Errorf("entry %s: %w", id, shared.ErrNotFound)
	}
	return participation.NewEntryFactory(nil).Reconstruct("e1", "r1", "Team A", []shared.MemberID{"m1", "m2"}), nil
}

func (entries storeRunningOrderEntries) ListByRace(ctx context.Context, _ shared.RaceID) ([]*participation.Entry, error) {
	entry, err := entries.FindByID(ctx, "e1")
	return []*participation.Entry{entry}, err
}

func (storeRunningOrderEntries) Save(context.Context, *participation.Entry) error { return nil }

type storeRunningOrders struct{ store *runningOrderStore }

func (orders storeRunningOrders) FindByEntry(_ context.Context, entryID shared.EntryID) (*runningorder.RunningOrder, error) {
	stored := orders.store.runningOrder
	if stored == nil || stored.EntryID() != entryID {
		return nil, fmt.Errorf("running order of %s: %w", entryID, shared.ErrNotFound)
	}
	return runningorder.NewFactory(nil).Reconstruct(stored.ID(), stored.EntryID(), stored.Status(), stored.Version(), assignmentsOf(stored)), nil
}

func (orders storeRunningOrders) Save(_ context.Context, saved *runningorder.RunningOrder) error {
	orders.store.runningOrder = runningorder.NewFactory(nil).Reconstruct(saved.ID(), saved.EntryID(), saved.Status(), saved.Version(), assignmentsOf(saved))
	return nil
}

func assignmentsOf(runningOrder *runningorder.RunningOrder) []runningorder.Assignment {
	var legs []runningorder.Assignment
	for _, leg := range runningOrder.Legs() {
		assignment := runningorder.Assignment{Number: leg.Number()}
		if runner, ok := leg.Runner(); ok {
			assignment.Runner = &runner
		}
		legs = append(legs, assignment)
	}
	return legs
}

type storeHistory struct{ store *runningOrderStore }

func (history storeHistory) Append(context.Context, *runningorder.HistoryEntry) error {
	history.store.history++
	return nil
}

type passLapRemover struct{}

func (passLapRemover) RemoveLaps(context.Context, shared.RaceID, []shared.MemberID) error { return nil }

type storeRunningOrderQueries struct{ store *runningOrderStore }

func (queries storeRunningOrderQueries) ListRunningOrders(_ context.Context, teamID shared.TeamID, raceID shared.RaceID, rosterMemberID *shared.MemberID) (runningorders.RaceRunningOrdersRecord, error) {
	if raceID != "r1" || teamID != runningOrderTeamID {
		return runningorders.RaceRunningOrdersRecord{}, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	record := runningorders.RaceRunningOrdersRecord{Description: runningOrderDescription(), Answer: race.NotInvited}
	if rosterMemberID != nil {
		record.Answer = queries.store.answer
	}
	if stored := queries.store.runningOrder; stored != nil {
		order := runningorders.RunningOrderRecord{EntryID: stored.EntryID(), EntryName: "Team A", Status: stored.Status(), Version: stored.Version()}
		for _, leg := range stored.Legs() {
			legRecord := runningorders.LegRecord{Number: leg.Number()}
			if runner, ok := leg.Runner(); ok {
				runnerRecord := runningOrderRunners[runner]
				legRecord.Runner = &runnerRecord
			}
			order.Legs = append(order.Legs, legRecord)
		}
		record.RunningOrders = []runningorders.RunningOrderRecord{order}
	}
	return record, nil
}

type passUnitOfWork struct{}

func (passUnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	return work(ctx)
}

func newRunningOrderRouter(store *runningOrderStore) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	now := func() time.Time { return time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC) }
	races, entries, orders, history := storeRunningOrderRaces{}, storeRunningOrderEntries{}, storeRunningOrders{store}, storeHistory{store}
	list := runningorders.NewListRunningOrdersUseCase(storeRunningOrderQueries{store}, race.VisibilityPolicy{}, runningorder.VisibilityPolicy{}, runningorders.Predictor{}, runningorders.RunningOrderDTOFactory{}, runningOrderTeamID)
	runningOrderController := controller.NewRunningOrderController(
		list,
		runningorders.NewGetRunningOrderUseCase(list),
		runningorders.NewSaveRunningOrderUseCase(passUnitOfWork{}, races, entries, orders, history, passLapRemover{}, runningorder.NewFactory(&sequentialIDs{}), runningorder.HistoryFactory{}, now, runningOrderTeamID),
		runningorders.NewConfirmRunningOrderUseCase(passUnitOfWork{}, races, entries, orders, history, runningorder.HistoryFactory{}, now, runningOrderTeamID),
		runningorders.RunningOrderCommandFactory{},
	)
	check := health.NewCheckUseCase(fakeHealth{}, fakeHealth{}, health.StatusDTOFactory{}, logger)
	return NewRouter(Controllers{Health: controller.NewHealthController(check), RunningOrders: runningOrderController}, Options{
		Authenticate:   middleware.Authenticate(tokenAuthenticator{}, logger),
		ResolveCaller:  middleware.ResolveCaller(tokenCallers{}, logger),
		FrontendOrigin: testOrigin,
		Logger:         logger,
	})
}

const firstDraftBody = `{"legs":[{"leg_number":1,"roster_member_id":"m1"},{"leg_number":2,"roster_member_id":null}]}`

func TestRunningOrderRoutesCallers(t *testing.T) {
	tests := []struct {
		method, path, body string
		store              func() *runningOrderStore
		want               map[string]int // by session token; "" is logged out
	}{
		{method: http.MethodGet, path: "/api/v1/races/r1/running-orders", store: func() *runningOrderStore { return newRunningOrderStore(race.Attending) },
			want: map[string]int{"": 401, waitingToken: 403, memberToken: 200, organizerToken: 200}},
		{method: http.MethodPut, path: "/api/v1/races/r1/entries/e1/running-order", body: firstDraftBody, store: func() *runningOrderStore { return newRunningOrderStore(race.Attending) },
			want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 200}},
		{method: http.MethodPost, path: "/api/v1/races/r1/entries/e1/running-order/confirm", body: `{"version":0}`,
			store: func() *runningOrderStore {
				return newRunningOrderStore(race.Attending).withOrder(runningorder.Draft, memberIDOf("m1"), memberIDOf("m2"))
			},
			want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 200}},
	}
	for _, testCase := range tests {
		for _, token := range []string{"", waitingToken, memberToken, organizerToken} {
			t.Run(fmt.Sprintf("%s %s as %q", testCase.method, testCase.path, token), func(t *testing.T) {
				rec := serve(newRunningOrderRouter(testCase.store()), testCase.method, testCase.path, testCase.body, sessionOf(token))
				if rec.Code != testCase.want[token] {
					t.Errorf("status = %d, want %d (body %s)", rec.Code, testCase.want[token], rec.Body)
				}
			})
		}
	}
}

func TestRunningOrderRoutesBodies(t *testing.T) {
	t.Run("returns the saved first draft with its prediction to the organizer", func(t *testing.T) {
		store := newRunningOrderStore(race.Attending)
		rec := serve(newRunningOrderRouter(store), http.MethodPut, "/api/v1/races/r1/entries/e1/running-order", firstDraftBody, sessionOf(organizerToken))
		want := `{"entry_id":"e1","entry_name":"Team A","legs":[{"distance_km":5,"leg_number":1,"runner":{"name":"Taro","roster_member_id":"m1"}},` +
			`{"distance_km":3.2,"leg_number":2,"runner":null}],"prediction":{"has_empty_legs":true,"laps":null,"total_time":"0:18:20"},"status":"draft","version":0}`
		if rec.Code != http.StatusOK || rec.Body.String() != want+"\n" {
			t.Errorf("status %d, body %s\nwant %s", rec.Code, rec.Body, want)
		}
		if store.history != 1 {
			t.Errorf("%d history entries, want 1", store.history)
		}
	})

	t.Run("shows an attending member the confirmed order without the version", func(t *testing.T) {
		store := newRunningOrderStore(race.Attending).withOrder(runningorder.Confirmed, memberIDOf("m1"), memberIDOf("m2"))
		rec := serve(newRunningOrderRouter(store), http.MethodGet, "/api/v1/races/r1/running-orders", "", sessionOf(memberToken))
		want := `{"running_orders":[{"entry_id":"e1","entry_name":"Team A","legs":[{"distance_km":5,"leg_number":1,"runner":{"name":"Taro","roster_member_id":"m1"}},` +
			`{"distance_km":3.2,"leg_number":2,"runner":{"name":"Jiro","roster_member_id":"m2"}}],"prediction":{"has_empty_legs":false,"laps":null,"total_time":"0:29:32"},"status":"confirmed"}]}`
		if rec.Code != http.StatusOK || rec.Body.String() != want+"\n" {
			t.Errorf("status %d, body %s\nwant %s", rec.Code, rec.Body, want)
		}
	})

	t.Run("shows an attending member no draft", func(t *testing.T) {
		store := newRunningOrderStore(race.Attending).withOrder(runningorder.Draft, memberIDOf("m1"), nil)
		rec := serve(newRunningOrderRouter(store), http.MethodGet, "/api/v1/races/r1/running-orders", "", sessionOf(memberToken))
		if want := `{"running_orders":[]}`; rec.Code != http.StatusOK || rec.Body.String() != want+"\n" {
			t.Errorf("status %d, body %s, want %s", rec.Code, rec.Body, want)
		}
	})
}

func TestRunningOrderRoutesErrors(t *testing.T) {
	draft := func() *runningOrderStore {
		return newRunningOrderStore(race.Attending).withOrder(runningorder.Draft, memberIDOf("m1"), nil)
	}
	confirmed := func() *runningOrderStore {
		return newRunningOrderStore(race.Attending).withOrder(runningorder.Confirmed, memberIDOf("m1"), memberIDOf("m2"))
	}
	tests := []struct {
		name, method, path, body, token string
		store                           func() *runningOrderStore
		wantStatus                      int
		wantCode                        string
	}{
		{name: "a member who may not see the race", method: http.MethodGet, path: "/api/v1/races/r1/running-orders", token: memberToken,
			store: func() *runningOrderStore { return newRunningOrderStore(race.NotAttending) }, wantStatus: 404, wantCode: "not_found"},
		{name: "a save without a body", method: http.MethodPut, path: "/api/v1/races/r1/entries/e1/running-order", token: organizerToken, store: draft, wantStatus: 400, wantCode: "invalid_request"},
		{name: "a version for an order not saved yet", method: http.MethodPut, path: "/api/v1/races/r1/entries/e1/running-order", body: `{"version":0,"legs":[{"leg_number":1},{"leg_number":2}]}`,
			token: organizerToken, store: func() *runningOrderStore { return newRunningOrderStore(race.Attending) }, wantStatus: 409, wantCode: "version_conflict"},
		{name: "a runner who is not in the entry", method: http.MethodPut, path: "/api/v1/races/r1/entries/e1/running-order", body: `{"version":0,"legs":[{"leg_number":1,"roster_member_id":"m9"},{"leg_number":2}]}`,
			token: organizerToken, store: draft, wantStatus: 422, wantCode: "validation_failed"},
		{name: "an entry that does not exist", method: http.MethodPut, path: "/api/v1/races/r1/entries/e9/running-order", body: firstDraftBody, token: organizerToken, store: draft, wantStatus: 404, wantCode: "not_found"},
		{name: "confirming with an empty leg", method: http.MethodPost, path: "/api/v1/races/r1/entries/e1/running-order/confirm", body: `{"version":0}`, token: organizerToken, store: draft, wantStatus: 422, wantCode: "validation_failed"},
		{name: "confirming a confirmed order", method: http.MethodPost, path: "/api/v1/races/r1/entries/e1/running-order/confirm", body: `{"version":0}`, token: organizerToken, store: confirmed, wantStatus: 409, wantCode: "state_conflict"},
		{name: "confirming with a stale version", method: http.MethodPost, path: "/api/v1/races/r1/entries/e1/running-order/confirm", body: `{"version":5}`, token: organizerToken, store: confirmed, wantStatus: 409, wantCode: "version_conflict"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := serve(newRunningOrderRouter(testCase.store()), testCase.method, testCase.path, testCase.body, sessionOf(testCase.token))
			if rec.Code != testCase.wantStatus || problemCode(t, rec.Body.Bytes()) != testCase.wantCode {
				t.Errorf("status %d, body %s, want %d %s", rec.Code, rec.Body, testCase.wantStatus, testCase.wantCode)
			}
		})
	}
}
