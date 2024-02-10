package bridgecmd

import (
	"github.com/spf13/cobra"

	"github.com/MichaelMure/git-bug/bridge"
	"github.com/MichaelMure/git-bug/commands/execenv"
)

func NewBridgeCommand(env *execenv.Env) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "bridge",
		Short:   "List bridges to other bug trackers",
		PreRunE: execenv.LoadBackend(env),
		RunE: execenv.CloseBackend(env, func(cmd *cobra.Command, args []string) error {
			return runBridge(env)
		}),
		Args: cobra.NoArgs,
	}

	subCmd, err := newBridgeAuthCommand(env)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(subCmd)
	subCmd, err = newBridgeNewCommand(env)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(subCmd)
	cmd.AddCommand(newBridgePullCommand(env))
	cmd.AddCommand(newBridgePushCommand(env))
	cmd.AddCommand(newBridgeRm(env))

	return cmd, nil
}

func runBridge(env *execenv.Env) error {
	configured, err := bridge.ConfiguredBridges(env.Backend)
	if err != nil {
		return err
	}

	for _, c := range configured {
		env.Out.Println(c)
	}

	return nil
}
