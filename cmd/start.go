package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-civo/pkg/civo"
	"github.com/spf13/cobra"
)

// StartCmd holds the cmd flags.
type StartCmd struct{}

// NewStartCmd defines a command.
func NewStartCmd() *cobra.Command {
	cmd := &StartCmd{}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			civoProvider, err := civo.NewProvider(true)
			if err != nil {
				return err
			}

			return cmd.Run(
				context.Background(),
				civoProvider,
			)
		},
	}

	return startCmd
}

// Run runs the command logic.
func (cmd *StartCmd) Run(
	ctx context.Context,
	providerCivo *civo.CivoProvider,
) error {
	return civo.Start(providerCivo)
}
