package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dhermes/tailsk8s/pkg/cli"
	"github.com/dhermes/tailsk8s/pkg/tailscale/command/authorize"
)

func run() error {
	ctx := context.Background()

	c, err := authorize.NewConfig()
	if err != nil {
		return err
	}
	debug := false
	cmd := &cobra.Command{
		Use:           "tailscale-authorize-device",
		Short:         "Authorize a new device to join a Tailnet",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := cli.WithDebug(ctx, debug)
			return authorize.AuthorizeDevice(ctx, c)
		},
	}

	cmd.PersistentFlags().StringVar(
		&c.APIConfig.Tailnet,
		"tailnet",
		c.APIConfig.Tailnet,
		"The Tailnet where the device exists",
	)
	cmd.PersistentFlags().StringVar(
		&c.APIConfig.APIKey,
		"api-key",
		c.APIConfig.APIKey,
		("The Tailscale API key; if it beings with \"file:\", then it will " +
			"be interpreted as a path to a file containing the Tailscale API key"),
	)
	cmd.PersistentFlags().StringVar(
		&c.Hostname,
		"hostname",
		c.Hostname,
		"The hostname of the device to authorize; if omitted the current device hostname will be used",
	)
	cmd.PersistentFlags().BoolVar(
		&debug,
		"debug",
		debug,
		"Enable extra print debugging",
	)

	required := []string{"tailnet", "api-key"}
	for _, name := range required {
		err := cobra.MarkFlagRequired(cmd.PersistentFlags(), name)
		if err != nil {
			return err
		}
	}

	return cmd.Execute()
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
