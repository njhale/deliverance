package memory

import (
	"context"

	"github.com/containerd/containerd/remotes"
	orascontent "github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/ecordell/bndlr/pkg/image"
	"github.com/ecordell/bndlr/pkg/registry/store"
)

type MemoryStore struct {
	store *orascontent.Memorystore
}

var _ store.Store = &MemoryStore{}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		store: orascontent.NewMemoryStore(),
	}
}

func (s *MemoryStore) Write(ctx context.Context, ref string, descriptor ocispec.Descriptor, blob []byte) error {
	s.store.Set(descriptor, blob)
	return nil
}

func (s *MemoryStore) Push(ctx context.Context, resolver remotes.Resolver, ref string, image *image.Descriptor) (*digest.Digest, error) {
	desc, err := oras.Push(ctx, resolver, ref, s.store, image.Layers, oras.WithConfig(image.Config), oras.WithManifest(image.Manifest), oras.WithNameValidation(nil))
	if err != nil {
		return nil, err
	}
	return &desc.Digest, nil
}
