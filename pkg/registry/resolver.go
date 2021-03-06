package registry

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	auth "github.com/deislabs/oras/pkg/auth/docker"
)

func NewResolver(username, password string, configs ...string) remotes.Resolver {
	if username != "" || password != "" {
		return docker.NewResolver(docker.ResolverOptions{
			Credentials: func(hostName string) (string, string, error) {
				return username, password, nil
			},
		})
	}
	cli, err := auth.NewClient(configs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error loading auth file: %v\n", err)
	}
	resolver, err := cli.Resolver(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error loading resolver: %v\n", err)
		resolver = docker.NewResolver(docker.ResolverOptions{})
	}
	return resolver
}

