package keepalive

// `sigi` -- the command, its flags, and everything it prints.

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
)

// compile-time proof that both actions satisfy Action.
var (
	_ Action = (*Mouse)(nil)
	_ Action = (*Keyboard)(nil)
)

// Version is injected at build time from the VERSION file via
// -ldflags "-X github.com/semirm-dev/sigi/internal/keepalive.Version=$(cat VERSION)".
var Version = "dev"

// Command builds the sigi command tree. There is one command, so this is it:
// flags and output live beside the loop they drive, and nothing is a
// package-level variable, so two of these can exist in one process.
func Command(out, errOut io.Writer) *cobra.Command {
	var (
		interval time.Duration
		verbose  bool
		name     string
	)

	cmd := &cobra.Command{
		Use:   "sigi",
		Short: "Keep this machine awake",
		Long: "sigi stops a machine going idle by doing something small on a timer:\n" +
			"nudging the mouse a pixel and putting it back, or tapping a key that\n" +
			"changes nothing.\n\n" +
			"It runs until you stop it with Ctrl-C.",
		Example: "  sigi                        # nudge the mouse every 2 minutes\n" +
			"  sigi --interval 30s         # more often\n" +
			"  sigi --action keyboard      # tap caps lock instead\n" +
			"  sigi --verbose              # say so on every tick",
		Version:       Version,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			action, err := newAction(name)
			if err != nil {
				return err
			}
			if verbose {
				action = reporting{Action: action, out: out}
			}

			fmt.Fprintf(out, "sigi: %s every %s, ctrl-c to stop\n", action.Name(), interval)

			if err := NewRunner(action, interval).Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Fprintln(out, "sigi: stopped")
			return nil
		},
	}

	cmd.Flags().DurationVarP(&interval, "interval", "i", DefaultInterval,
		"How long to wait between actions, e.g. 30s or 2m")
	cmd.Flags().StringVarP(&name, "action", "a", "mouse",
		"What to do on each tick: mouse or keyboard")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"Print each action as it happens")

	return cmd
}

// newAction resolves the --action flag. Naming the choices in the error beats
// leaving someone to guess which words are accepted.
func newAction(name string) (Action, error) {
	switch name {
	case "mouse":
		return NewMouse()
	case "keyboard":
		return NewKeyboard()
	default:
		return nil, fmt.Errorf("unknown action %q: choose mouse or keyboard", name)
	}
}
