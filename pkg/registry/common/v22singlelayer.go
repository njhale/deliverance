package common

import (
	"context"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"

	"github.com/ecordell/bndlr/pkg/image/builder"
	"github.com/ecordell/bndlr/pkg/image/layer"
	"github.com/ecordell/bndlr/pkg/registry/store"
)

// This package contains aggregate functions that wire together common options exposed by underlying components

// BuildAndPushDirectoryV22 builds and pushes a minimal v2-2 image with single layer built from a directory
func BuildAndPushDirectoryV22(ctx context.Context, ref string, s store.Store, resolver remotes.Resolver, dir string) (*digest.Digest, error) {
	builder, err := builder.NewMinimalV22Builder()
	if err != nil {
		return nil, err
	}

	l, err := layer.LayerFromDirectory(dir, layer.WithMediaType(images.MediaTypeDockerSchema2LayerGzip))
	if err != nil {
		return nil, err
	}

	image, err := builder.BuildImage(ctx, ref, s, layer.Layers{*l})
	if err != nil {
		return nil, err
	}

	return s.Push(ctx, resolver, ref, image)
}
