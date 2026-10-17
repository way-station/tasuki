package race

// Answer is the viewer's attendance answer for a race, as far as visibility is concerned.
type Answer int

// Answers.
const (
	NotInvited Answer = iota
	NotAnswered
	Undecided
	Attending
	NotAttending
)

// Viewer is who is looking at a race.
type Viewer struct {
	Organizer bool
	Answer    Answer // for organizers too, when they are linked to a roster member
}

// Visible is what a viewer may see of a race (ADR-0006, docs/design/database.md).
type Visible struct {
	Race                bool // the race at all: its name and date in the list, and its detail
	Description         bool
	Deadline            bool
	RunningOrderSummary bool // whether the running order is confirmed
}

// VisibilityPolicy decides what a viewer may see of a race.
type VisibilityPolicy struct{}

// For returns what the viewer may see.
// Members see only races they are invited to, until they answer "not attending".
func (VisibilityPolicy) For(v Viewer) Visible {
	if v.Organizer {
		return Visible{Race: true, Description: true, Deadline: true, RunningOrderSummary: true}
	}
	switch v.Answer {
	case NotAnswered, Undecided:
		return Visible{Race: true, Description: true}
	case Attending:
		return Visible{Race: true, Description: true, RunningOrderSummary: true}
	}
	return Visible{}
}
