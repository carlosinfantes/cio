// Package commands implements the interactive REPL mode.
package commands

import (
	"fmt"

	"github.com/carlosinfantes/cto-advisory-board/internal/cli/output"
	"github.com/carlosinfantes/cto-advisory-board/internal/cli/repl"
)

// RunInteractive starts the interactive REPL mode.
func RunInteractive() error {
	r, err := repl.New()
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to start REPL: %v", err))
		return err
	}
	defer r.Close()

	return r.Run()
}
