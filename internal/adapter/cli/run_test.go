package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRootHelpListsTopLevelCommands(t *testing.T) {
	t.Parallel()

	root := newRootCommand()
	root.SetArgs([]string{"--help"})
	var out bytes.Buffer
	root.SetOut(&out)

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	help := out.String()
	for _, name := range []string{"server", "policy", "storage", "client", "user", "scope", "consent", "purge"} {
		if !strings.Contains(help, name) {
			t.Errorf("root help missing %q:\n%s", name, help)
		}
	}
}

func TestLeafReturnsNotImplemented(t *testing.T) {
	t.Parallel()

	cases := [][]string{
		{"server"},
		{"policy", "init"},
		{"client", "create"},
		{"client", "deactivate"},
		{"client", "reactivate"},
		{"client", "allow"},
		{"client", "disallow"},
		{"client", "set-not-before"},
		{"user", "deactivate"},
		{"user", "reactivate"},
		{"user", "set-password"},
		{"user", "set-not-before"},
		{"scope", "deactivate"},
		{"scope", "reactivate"},
		{"scope", "list"},
		{"consent", "deactivate"},
		{"consent", "reactivate"},
		{"purge"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()
			err := Run(context.Background(), append([]string{"oauth"}, args...))
			if !errors.Is(err, ErrNotImplemented) {
				t.Fatalf("Run(%v) error = %v, want ErrNotImplemented", args, err)
			}
		})
	}
}
