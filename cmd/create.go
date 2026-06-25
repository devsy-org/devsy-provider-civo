package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-civo/pkg/civo"
	"github.com/spf13/cobra"
)

// CreateCmd holds the cmd flags.
type CreateCmd struct{}

// NewCreateCmd defines a command.
func NewCreateCmd() *cobra.Command {
	cmd := &CreateCmd{}
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an instance",
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

	return createCmd
}

// Run runs the command logic.
func (cmd *CreateCmd) Run(
	ctx context.Context,
	providerCivo *civo.CivoProvider,
) error {
	return civo.Create(providerCivo)
}
