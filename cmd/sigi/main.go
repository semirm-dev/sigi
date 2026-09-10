// Command sigi keeps a machine awake.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/semirm-dev/sigi/internal/keepalive"
)

func main() {
	// NotifyContext cancels the context the runner is watching, so ctrl-c
	// stops the loop rather than leaving it mid-tick while the process exits.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd := keepalive.Command(os.Stdout, os.Stderr)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "sigi:", err)
		os.Exit(1)
	}
}
