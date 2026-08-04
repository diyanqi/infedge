// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestServerCommandsRunMigrationsBeforeStart(t *testing.T) {
	t.Helper()

	if rootCmd.PreRun == nil {
		t.Fatal("root command must migrate before starting the default all mode")
	}

	for _, command := range []*struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "all", cmd: allCmd},
		{name: "api", cmd: apiCmd},
		{name: "worker", cmd: workerCmd},
		{name: "scheduler", cmd: schedulerCmd},
	} {
		if command.cmd == nil {
			t.Fatalf("%s command is nil", command.name)
		}
		if command.cmd.Name() != command.name {
			t.Fatalf("command name = %q, want %q", command.cmd.Name(), command.name)
		}
		// The hook invokes both relational and ClickHouse migrations before Run.
		if command.cmd.PreRun == nil {
			t.Fatalf("%s command must migrate before starting", command.name)
		}
	}
}
