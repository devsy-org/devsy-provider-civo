package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"

	"github.com/devsy-org/devsy/logging"
	"github.com/devsy-org/devsy/logging/redact"
	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// NewRootCmd returns a new root command.
func NewRootCmd() *cobra.Command {
	civoCmd := &cobra.Command{
		Use:           "devsy-provider-civo",
		Short:         "civo Provider commands",
		SilenceErrors: true,
		SilenceUsage:  true,

		PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
			var token struct {
				APIKey string `json:"apikey"`
			}
			_ = json.Unmarshal([]byte(os.Getenv("CIVO_TOKEN")), &token)
			redactor := redact.FromEntries([]string{"CIVO_API_KEY=" + token.APIKey})
			invalidConfig := func(name string) error {
				log.Init(
					log.Config{
						Level:        log.LevelInfoName,
						DefaultLevel: log.LevelInfoName,
						Format:       "json",
						Redactor:     redactor,
					},
				)
				return errors.New("invalid " + name)
			}

			cfg, invalid := logging.ResolveProvider(
				os.Getenv("DEVSY_LOG_LEVEL"),
				os.Getenv("DEVSY_LOG_FORMAT"),
			)
			if invalid.Level {
				return invalidConfig("DEVSY_LOG_LEVEL")
			}
			if invalid.Format {
				return invalidConfig("DEVSY_LOG_FORMAT")
			}

			log.Init(
				log.Config{
					Level:        logging.LevelName(cfg.Level),
					DefaultLevel: log.LevelInfoName,
					Format:       string(cfg.Format),
					Redactor:     redactor,
				},
			)

			return nil
		},
	}

	return civoCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// build the root command
	rootCmd := BuildRoot()

	// execute command
	err := rootCmd.Execute()
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			os.Exit(exitErr.ExitStatus())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
				log.Error(string(exitErr.Stderr))
			}
			os.Exit(exitErr.ExitCode())
		}

		log.Fatal(err)
	}
}

// BuildRoot creates a new root command with all subcommands attached.
func BuildRoot() *cobra.Command {
	rootCmd := NewRootCmd()

	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewDeleteCmd())
	rootCmd.AddCommand(NewCommandCmd())
	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewStopCmd())
	rootCmd.AddCommand(NewStatusCmd())
	rootCmd.AddCommand(NewTokenCmd())
	return rootCmd
}
