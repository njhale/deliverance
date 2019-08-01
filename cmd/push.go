package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	auth "github.com/deislabs/oras/pkg/auth/docker"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ecordell/bndlr/pkg/bundle"
	"github.com/ecordell/bndlr/pkg/signals"
)

type pushOptions struct {
	configs  []string
}

var pushOpts pushOptions

func newResolver(configs ...string) remotes.Resolver {
	client, err := auth.NewClient(configs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error loading auth file: %v\n", err)
	}
	resolver, err := client.Resolver(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error loading resolver: %v\n", err)
		resolver = docker.NewResolver(docker.ResolverOptions{})
	}
	return resolver
}

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signals.Context()

		if len(args) < 2  {
			return fmt.Errorf("should be called with two args: dir host")
		}
		dir := args[0]
		host := args[1]
		logrus.SetLevel(logrus.DebugLevel)
		resolver := newResolver(pushOpts.configs...)

		return bundle.PushDir(ctx, resolver, host, dir)
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringArrayVarP(&pushOpts.configs, "config", "c", []string{"~/.docker/config.json"}, "auth config path")
}
