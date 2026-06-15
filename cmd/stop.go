package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-civo/pkg/civo"
	"github.com/devsy-org/log"
	"github.com/spf13/cobra"
)

// StopCmd holds the cmd flags.
type StopCmd struct{}

// NewStopCmd defines a command.
func NewStopCmd() *cobra.Command {
	cmd := &StopCmd{}
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			civoProvider, err := civo.NewProvider(false, log.Default)
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

	return stopCmd
}

// Run runs the command logic.
func (cmd *StopCmd) Run(
	ctx context.Context,
	providerCivo *civo.CivoProvider,
	logs log.Logger,
) error {
	return civo.Stop(providerCivo)
}
