package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-civo/pkg/civo"
	"github.com/devsy-org/log"
	"github.com/spf13/cobra"
)

// DeleteCmd holds the cmd flags.
type DeleteCmd struct{}

// NewDeleteCmd defines a command.
func NewDeleteCmd() *cobra.Command {
	cmd := &DeleteCmd{}
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an instance",
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

	return deleteCmd
}

// Run runs the command logic.
func (cmd *DeleteCmd) Run(
	ctx context.Context,
	providerCivo *civo.CivoProvider,
	logs log.Logger,
) error {
	return civo.Delete(providerCivo)
}
