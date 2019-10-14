package store

import (
	"context"

	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/ecordell/bndlr/pkg/image"
)

type Store interface {
	// Write writes a descriptor and blob into the store
	Write(ctx context.Context, ref string, descriptor ocispec.Descriptor, blob []byte) error

	// Push takes a config, a manifest, and a set of layer descriptors and pushes it to the remote
	// fetching the blobs to push requires knowledge of the backing store, which is why this method is on the Store
	Push(ctx context.Context, resolver remotes.Resolver, ref string, image *image.Descriptor) (*digest.Digest, error)
}
