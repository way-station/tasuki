package config

import "testing"

// Deliberately failing, to check that main's ruleset blocks a pull request with failing CI (#37).
// This pull request is closed without merging.
func TestRulesetBlocksFailingCI(t *testing.T) {
	t.Fatal("deliberate failure")
}
