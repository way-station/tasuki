package rosters

import (
	"context"
	"testing"
)

func TestListPendingLinksUseCase(t *testing.T) {
	query := &fakeRosterQueryService{pending: []PendingLinkRecord{{LinkID: "k1", LineDisplayName: "taro", MemberID: "m1", MemberName: "Taro", CreatedAt: testNow}}}
	links, err := NewListPendingLinksUseCase(query, RosterDTOFactory{}, teamID).Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(links) != 1 || links[0].LinkID() != "k1" || links[0].LineDisplayName() != "taro" || links[0].MemberName() != "Taro" || !links[0].CreatedAt().Equal(testNow) {
		t.Errorf("Execute() = %+v, want the link of Taro", links)
	}
}
