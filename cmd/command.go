package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/devsy-org/devsy-provider-civo/pkg/civo"
	"github.com/devsy-org/devsy/pkg/ssh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CommandCmd holds the cmd flags.
type CommandCmd struct{}

// NewCommandCmd defines a command.
func NewCommandCmd() *cobra.Command {
	cmd := &CommandCmd{}
	commandCmd := &cobra.Command{
		Use:   "command",
		Short: "Command an instance",
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

	return commandCmd
}

// Run runs the command logic.
func (cmd *CommandCmd) Run(
	ctx context.Context,
	providerCivo *civo.CivoProvider,
) error {
	command := os.Getenv("COMMAND")

	if command == "" {
		return fmt.Errorf("command environment variable is missing")
	}

	// get instance
	instance, err := civo.GetDevsyInstance(providerCivo)
	if err != nil {
		return errors.Wrap(err, "get instance")
	}

	sshClient, err := ssh.NewSSHPassClient(
		"civo",
		instance.PublicIP+":22",
		instance.InitialPassword,
	)
	if err != nil {
		return errors.Wrap(err, "create ssh client")
	}

	defer func() { _ = sshClient.Close() }()

	// run command
	return ssh.Run(ctx, ssh.RunOptions{
		Client:  sshClient,
		Command: command,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
}
