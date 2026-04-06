package main

import "testing"

func TestParserAcceptsSinglePercentValues(t *testing.T) {
	pdx = nil
	if err := ensureParser(); err != nil {
		t.Fatalf("ensureParser: %v", err)
	}

	input := `focus = { completion_reward = { war_support = 35% } }`
	if _, err := parsePdxSource(input); err != nil {
		t.Fatalf("parse single percent value: %v", err)
	}
}