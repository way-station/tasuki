package participations

import (
	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const teamID shared.TeamID = "t1"

// member returns a caller linked to roster member m1.
func member() auth.MeDTO {
	factory := auth.MeDTOFactory{}
	return factory.Create("u1", "Taro", false, identity.CallerApproved, factory.CreateRosterMember("m1", "Taro"))
}

// organizer returns an organizer without a link.
func organizer() auth.MeDTO {
	return auth.MeDTOFactory{}.Create("u2", "Hanako", true, identity.CallerNone, nil)
}
