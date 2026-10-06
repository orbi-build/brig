package main

import (
	"slices"
	"testing"

	"github.com/brig-sh/brig/internal/profile"
	"github.com/brig-sh/brig/internal/wrap"
)

// agentArgs adds the agent's own session-name flag to the arguments handed to
// the agent CLI. Only claude-code takes the session label as its display name;
// every other profile must get the tail unchanged, or brig would hand an agent
// a flag it does not read. The table is built from the embedded profiles, so a
// profile added later is covered without naming it here.
func TestAgentArgsOnlyNamesClaudeCode(t *testing.T) {
	tail := []string{"-p", "hi"}

	type testCase struct {
		name    string
		profile string
		rawName string
		want    []string
	}

	cases := []testCase{
		{"claude-code takes the label as its display name", "claude-code", "refactor", []string{"--name", "refactor", "-p", "hi"}},
		{"claude-code without a label leaves the tail alone", "claude-code", "", []string{"-p", "hi"}},
	}

	others := 0
	for _, p := range profile.All() {
		if !profile.Embedded(p.Name) || p.Name == "claude-code" {
			continue
		}
		others++
		cases = append(cases, testCase{p.Name + " gets no session-name flag", p.Name, "refactor", []string{"-p", "hi"}})
	}
	if others == 0 {
		t.Fatal("no embedded profile besides claude-code. The table would pass vacuously")
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := agentArgs(&wrap.Config{RawName: c.rawName}, profile.Profile{Name: c.profile}, tail)
			if !slices.Equal(got, c.want) {
				t.Fatalf("agentArgs(%s, %q) = %q, want %q", c.profile, c.rawName, got, c.want)
			}
			if c.profile != "claude-code" && slices.Contains(got, "--name") {
				t.Errorf("agentArgs(%s) passed --name; only claude-code takes it", c.profile)
			}
		})
	}
}
