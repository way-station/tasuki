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

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/application/races"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

const raceTeamID shared.TeamID = "team-1"

// The session token of each caller in these tests is their user ID.
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
		return factory.Create(userID, "Member", false, identity.CallerApproved, factory.CreateRosterMember("m1", "Member")), nil
	}
	return factory.Create(userID, "Waiting", false, identity.CallerPending, nil), nil
}

// memoryRaces stores races in memory and reads them as the race query service does.
// answers holds the answers of roster member m1.
type memoryRaces struct {
	saved   []*race.Race
	answers map[shared.RaceID]race.Answer
}

func (store *memoryRaces) FindByID(_ context.Context, id shared.RaceID) (*race.Race, error) {
	for _, saved := range store.saved {
		if saved.ID() == id {
			return race.NewFactory(nil).Reconstruct(saved.ID(), saved.TeamID(), saved.Name(), saved.Description(), saved.Version()), nil
		}
	}
	return nil, fmt.Errorf("race %s: %w", id, shared.ErrNotFound)
}

func (store *memoryRaces) Save(_ context.Context, saved *race.Race) error {
	for index, existing := range store.saved {
		if existing.ID() == saved.ID() {
			if existing.Version() != saved.LoadedVersion() {
				return shared.ErrVersionConflict
			}
			store.saved[index] = saved
			return nil
		}
	}
	store.saved = append(store.saved, saved)
	return nil
}

func (store *memoryRaces) answerOf(raceID shared.RaceID, rosterMemberID *shared.MemberID) race.Answer {
	if rosterMemberID == nil {
		return race.NotInvited
	}
	return store.answers[raceID]
}

func (store *memoryRaces) ListRaces(_ context.Context, teamID shared.TeamID, rosterMemberID *shared.MemberID) ([]races.RaceSummaryRecord, error) {
	var records []races.RaceSummaryRecord
	for _, saved := range slices.Backward(store.saved) {
		if saved.TeamID() == teamID {
			records = append(records, races.RaceSummaryRecord{ID: saved.ID(), Name: saved.Name(), RaceDate: saved.Description().RaceDate(), Answer: store.answerOf(saved.ID(), rosterMemberID)})
		}
	}
	return records, nil
}

func (store *memoryRaces) FindRace(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID, rosterMemberID *shared.MemberID) (races.RaceRecord, error) {
	found, err := store.FindByID(ctx, raceID)
	if err != nil || found.TeamID() != teamID {
		return races.RaceRecord{}, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	description := found.Description()
	target, hasTarget := description.TargetTotalTime()
	params := race.DescriptionParams{
		Type: description.Type(), RaceDate: description.RaceDate(), Deadline: description.Deadline(), Text: description.Text(),
		LegDistancesKm: description.LegDistancesKm(), TimeLimitMinutes: description.TimeLimitMinutes(), LapDistanceKm: description.LapDistanceKm(),
	}
	if hasTarget {
		params.TargetTotalTime = &target
	}
	if laps, ok := description.TargetLaps(); ok {
		params.TargetLaps = &laps
	}
	return races.RaceRecord{ID: raceID, Name: found.Name(), Version: found.Version(), Description: params, Answer: store.answerOf(raceID, rosterMemberID)}, nil
}

type passUnitOfWork struct{}

func (passUnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	return work(ctx)
}

type passAdjuster struct{}

func (passAdjuster) AdjustRunningOrders(context.Context, shared.RaceID, race.Description) error {
	return nil
}

func newRaceRouter(store *memoryRaces) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	factory := race.NewFactory(&sequentialIDs{})
	raceController := controller.NewRaceController(
		races.NewListRacesUseCase(store, race.VisibilityPolicy{}, races.RaceDTOFactory{}, raceTeamID),
		races.NewGetRaceUseCase(store, race.VisibilityPolicy{}, races.RaceDTOFactory{}, raceTeamID),
		races.NewCreateRaceUseCase(store, factory, raceTeamID),
		races.NewUpdateRaceUseCase(passUnitOfWork{}, store, passAdjuster{}, raceTeamID),
		races.RaceCommandFactory{},
	)
	check := health.NewCheckUseCase(fakeHealth{}, fakeHealth{}, health.StatusDTOFactory{}, logger)
	return NewRouter(Controllers{Health: controller.NewHealthController(check), Races: raceController}, Options{
		Authenticate:   middleware.Authenticate(tokenAuthenticator{}, logger),
		ResolveCaller:  middleware.ResolveCaller(tokenCallers{}, logger),
		FrontendOrigin: testOrigin,
		Logger:         logger,
	})
}

func sessionOf(token string) *http.Cookie {
	if token == "" {
		return nil
	}
	return &http.Cookie{Name: middleware.SessionCookieName, Value: token, Path: "/api", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode}
}

const distanceRaceBody = `{"name":"Hakone","description":{"type":"distance","race_date":"2026-11-23",` +
	`"deadline":"2026-11-01T18:00:00+09:00","text":"Meet at 7:00","leg_distances_km":[5,3.2],"target":{"total_time":"2:45:00"}}}`

func decodeObject(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var object map[string]any
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return object
}

func problemCode(t *testing.T, body []byte) string {
	t.Helper()
	code, _ := decodeObject(t, body)["code"].(string)
	return code
}

func createTestRace(t *testing.T, handler http.Handler) string {
	t.Helper()
	rec := serve(handler, http.MethodPost, "/api/v1/races", distanceRaceBody, sessionOf(organizerToken))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /races status = %d, body %s", rec.Code, rec.Body)
	}
	id, _ := decodeObject(t, rec.Body.Bytes())["id"].(string)
	return id
}

func TestRaceRoutesCallers(t *testing.T) {
	handler := newRaceRouter(&memoryRaces{})
	raceID := createTestRace(t, handler)

	tests := []struct {
		method, path, body string
		want               map[string]int // by session token; "" is logged out
	}{
		{method: http.MethodGet, path: "/api/v1/races", want: map[string]int{"": 401, waitingToken: 403, memberToken: 200, organizerToken: 200}},
		{method: http.MethodPost, path: "/api/v1/races", body: distanceRaceBody, want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 201}},
		{method: http.MethodGet, path: "/api/v1/races/" + raceID, want: map[string]int{"": 401, waitingToken: 403, memberToken: 404, organizerToken: 200}},
		{method: http.MethodPut, path: "/api/v1/races/" + raceID, body: `{"version":0,` + distanceRaceBody[1:], want: map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: 200}},
	}
	for _, testCase := range tests {
		for _, token := range []string{"", waitingToken, memberToken, organizerToken} {
			t.Run(fmt.Sprintf("%s %s as %q", testCase.method, testCase.path, token), func(t *testing.T) {
				rec := serve(handler, testCase.method, testCase.path, testCase.body, sessionOf(token))
				if rec.Code != testCase.want[token] {
					t.Errorf("status = %d, want %d (body %s)", rec.Code, testCase.want[token], rec.Body)
				}
			})
		}
	}
}

func TestRaceRoutesBodies(t *testing.T) {
	t.Run("returns the created race to the organizer", func(t *testing.T) {
		handler := newRaceRouter(&memoryRaces{})
		rec := serve(handler, http.MethodPost, "/api/v1/races", distanceRaceBody, sessionOf(organizerToken))
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
		}
		got := decodeObject(t, rec.Body.Bytes())
		want := `{"can_submit_laps":false,"description":{"deadline":"2026-11-01T18:00:00+09:00","leg_distances_km":[5,3.2],"race_date":"2026-11-23",` +
			`"target":{"total_time":"2:45:00"},"text":"Meet at 7:00","type":"distance"},"id":"id1","name":"Hakone","running_order_confirmed":false,"version":0}`
		if body, _ := json.Marshal(got); string(body) != want {
			t.Errorf("body = %s\nwant   %s", body, want)
		}
	})

	t.Run("shows a member null for an invitation without an answer, and no deadline or version", func(t *testing.T) {
		store := &memoryRaces{answers: map[shared.RaceID]race.Answer{}}
		handler := newRaceRouter(store)
		raceID := createTestRace(t, handler)
		store.answers[shared.RaceID(raceID)] = race.NotAnswered

		rec := serve(handler, http.MethodGet, "/api/v1/races/"+raceID, "", sessionOf(memberToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
		}
		got := decodeObject(t, rec.Body.Bytes())
		if attendance, ok := got["my_attendance"]; !ok || attendance != nil {
			t.Errorf("my_attendance = %v (present %v), want null", attendance, ok)
		}
		description, _ := got["description"].(map[string]any)
		for _, hidden := range []bool{hasKey(got, "version"), hasKey(got, "running_order_confirmed"), hasKey(description, "deadline")} {
			if hidden {
				t.Errorf("body %s shows the member what only organizers see", rec.Body)
			}
		}

		list := serve(handler, http.MethodGet, "/api/v1/races", "", sessionOf(memberToken))
		races, _ := decodeObject(t, list.Body.Bytes())["races"].([]any)
		if len(races) != 1 {
			t.Fatalf("GET /races = %s, want the one invitation", list.Body)
		}
		if summary, _ := races[0].(map[string]any); !hasKey(summary, "my_attendance") || summary["my_attendance"] != nil {
			t.Errorf("summary = %v, want my_attendance null", summary)
		}
	})

	t.Run("omits my_attendance for an organizer who is not invited", func(t *testing.T) {
		handler := newRaceRouter(&memoryRaces{})
		raceID := createTestRace(t, handler)
		rec := serve(handler, http.MethodGet, "/api/v1/races/"+raceID, "", sessionOf(organizerToken))
		if hasKey(decodeObject(t, rec.Body.Bytes()), "my_attendance") {
			t.Errorf("body = %s, want no my_attendance", rec.Body)
		}
	})
}

func TestRaceRoutesErrors(t *testing.T) {
	handler := newRaceRouter(&memoryRaces{})
	raceID := createTestRace(t, handler)
	update := func(version int) string { return fmt.Sprintf(`{"version":%d,`, version) + distanceRaceBody[1:] }
	noLegs := `{"name":"Hakone","description":{"type":"distance","race_date":"2026-11-23","deadline":"2026-11-01T18:00:00+09:00"}}`

	tests := []struct {
		name, method, path, body string
		status                   int
		code                     string
	}{
		{name: "a malformed body", method: http.MethodPost, path: "/api/v1/races", body: `{"name":`, status: 400, code: "invalid_request"},
		{name: "a distance race without legs", method: http.MethodPost, path: "/api/v1/races", body: noLegs, status: 422, code: "validation_failed"},
		{name: "a race that does not exist", method: http.MethodGet, path: "/api/v1/races/unknown", status: 404, code: "not_found"},
		{name: "editing a race that does not exist", method: http.MethodPut, path: "/api/v1/races/unknown", body: update(0), status: 404, code: "not_found"},
		{name: "editing with a stale version", method: http.MethodPut, path: "/api/v1/races/" + raceID, body: update(5), status: 409, code: "version_conflict"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := serve(handler, testCase.method, testCase.path, testCase.body, sessionOf(organizerToken))
			if rec.Code != testCase.status || problemCode(t, rec.Body.Bytes()) != testCase.code {
				t.Errorf("status %d, body %s; want %d %s", rec.Code, rec.Body, testCase.status, testCase.code)
			}
		})
	}

	t.Run("lists the field that breaks a rule", func(t *testing.T) {
		rec := serve(handler, http.MethodPost, "/api/v1/races", noLegs, sessionOf(organizerToken))
		errorsField, _ := decodeObject(t, rec.Body.Bytes())["errors"].([]any)
		if len(errorsField) != 1 || errorsField[0].(map[string]any)["field"] != "leg_distances_km" {
			t.Errorf("errors = %v, want leg_distances_km", errorsField)
		}
	})
}

func hasKey(object map[string]any, key string) bool {
	_, ok := object[key]
	return ok
}
