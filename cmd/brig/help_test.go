package main

import (
	"strings"
	"testing"

	"github.com/brig-sh/brig/internal/runtime"
)

// The top-level verbs the issue lists. Before this change, each refused --help
// and -h with a usage error and printed nothing to stdout.
var topLevelVerbs = []string{"run", "sh", "info", "rm", "stop", "ls", "doctor", "version"}

// Asking a verb for help is a question, not a mistake: it is answered with that
// verb's usage, on stdout, and the exit status is 0.
func TestEveryTopLevelVerbAnswersHelp(t *testing.T) {
	for _, verb := range topLevelVerbs {
		for _, flag := range []string{"--help", "-h"} {
			out, err := captureStdout(t, func() error { return run([]string{verb, flag}) })
			if err != nil {
				t.Errorf("brig %s %s: %v", verb, flag, err)
				continue
			}
			if out == "" {
				t.Errorf("brig %s %s printed nothing", verb, flag)
			}
			if out == usage {
				t.Errorf("brig %s %s printed the global usage, not the verb's own:\n%s", verb, flag, out)
			}
			if !strings.Contains(out, "brig "+verb) {
				t.Errorf("brig %s %s does not name the verb:\n%s", verb, flag, out)
			}
		}
	}
}

// The two spellings of the same question answer with the same text, so a reader
// who learns one has learnt the other.
func TestHelpVerbPrintsTheTextTheFlagDoes(t *testing.T) {
	t.Setenv("BRIG_PROFILE_DIR", t.TempDir())
	for _, verb := range topLevelVerbs {
		byFlag, err := captureStdout(t, func() error { return run([]string{verb, "--help"}) })
		if err != nil {
			t.Fatalf("brig %s --help: %v", verb, err)
		}
		byVerb, err := captureStdout(t, func() error { return run([]string{"help", verb}) })
		if err != nil {
			t.Fatalf("brig help %s: %v", verb, err)
		}
		if byFlag != byVerb {
			t.Errorf("brig help %s and brig %s --help differ:\n--- brig help %s ---\n%s\n--- brig %s --help ---\n%s",
				verb, verb, verb, byVerb, verb, byFlag)
		}
	}
}

// A verb group is a verb too: `brig help agent` answers with the group's own
// usage, the text `brig agent --help` prints, which is unchanged.
func TestHelpNamesAVerbGroup(t *testing.T) {
	t.Setenv("BRIG_PROFILE_DIR", t.TempDir())
	for _, verb := range []string{"agent", "secret", "policy", "network", "telemetry", "completion"} {
		byFlag, err := captureStdout(t, func() error { return run([]string{verb, "--help"}) })
		if err != nil {
			t.Fatalf("brig %s --help: %v", verb, err)
		}
		byVerb, err := captureStdout(t, func() error { return run([]string{"help", verb}) })
		if err != nil {
			t.Fatalf("brig help %s: %v", verb, err)
		}
		if byFlag != byVerb {
			t.Errorf("brig help %s and brig %s --help differ:\n--- brig help %s ---\n%s\n--- brig %s --help ---\n%s",
				verb, verb, verb, byVerb, verb, byFlag)
		}
	}
}

// A bare help request is the command list, as it always was, and an unknown
// word keeps that answer rather than a refusal.
func TestBareHelpStillPrintsTheGlobalUsage(t *testing.T) {
	t.Setenv("BRIG_PROFILE_DIR", t.TempDir())
	for _, args := range [][]string{{"help"}, {"help", "nosuchverb"}} {
		out, err := captureStdout(t, func() error { return run(args) })
		if err != nil {
			t.Errorf("brig %v: %v", args, err)
			continue
		}
		if out != usage {
			t.Errorf("brig %v is no longer the global usage:\n%s", args, out)
		}
	}
}

// --help after the ref belongs to the agent: the ref is already brig's, so the
// next --help is the agent's word and split must hand it over untouched. This
// is the line the two spellings above must not break.
func TestParseKeepsHelpAfterProfileForTheAgent(t *testing.T) {
	_, name, tail, err := parse("run", []string{"claude", "--help"})
	if err != nil {
		t.Fatalf("parse run claude --help: %v", err)
	}
	if name != "claude" || strings.Join(tail, " ") != "--help" {
		t.Errorf("parse = (%q, %q), want (claude, --help)", name, tail)
	}
}

// And through the whole run: the fake agent is started with --help in its argv,
// rather than brig answering it as its own.
func TestRunForwardsHelpAfterTheRefToTheAgent(t *testing.T) {
	var argv []string
	rt := &jsonRuntime{attachFn: func(spec runtime.ExecSpec) (int, error) {
		argv = append([]string(nil), spec.Cmd...)
		return 0, nil
	}}
	jsonRunHost(t, rt)

	// Under --json a successful agent run comes back as an agentExit carrying
	// its status, so 0 is the success here the way main reads it.
	if _, err := captureStdout(t, func() error { return run([]string{"--json", "run", "faker", "--help"}) }); exitCode(err) != 0 {
		t.Fatalf("run faker --help: exit %d", exitCode(err))
	}
	if len(argv) == 0 || argv[len(argv)-1] != "--help" {
		t.Errorf("agent argv = %q, want --help forwarded", argv)
	}
}
