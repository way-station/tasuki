package rosters

import "testing"

func TestRosterCommandFactory(t *testing.T) {
	if command := (RosterCommandFactory{}).AddRosterMember("Taro"); command.name != "Taro" {
		t.Errorf("AddRosterMember() = %+v, want Taro", command)
	}
	if command := (RosterCommandFactory{}).SetLeft("m1", true); command.memberID != "m1" || !command.left {
		t.Errorf("SetLeft() = %+v, want m1 left", command)
	}
}
