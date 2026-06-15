package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-civo/pkg/civo"
	"github.com/devsy-org/log"
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
			civoProvider, err := civo.NewProvider(true, log.Default)
			if err != nil {
				return err
			}

			return cmd.Run(
				context.Background(),
				civoProvider,
				log.Default,
			)
		},
	}

	return startCmd
}

// Run runs the command logic.
func (cmd *StartCmd) Run(
	ctx context.Context,
	providerCivo *civo.CivoProvider,
	logs log.Logger,
) error {
	return civo.Start(providerCivo)
}
