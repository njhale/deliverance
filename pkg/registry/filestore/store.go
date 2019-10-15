package filestore

import (
	"context"
	"io/ioutil"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/ecordell/bndlr/pkg/image"
	"github.com/ecordell/bndlr/pkg/registry/store"
)

type FileStore struct {
	store    content.Store
}

var _ store.Store = &FileStore{}

func NewTmpFileStore() (*FileStore, error) {
	tmpdir, err := ioutil.TempDir("", "bndlr-")
	if err != nil {
		return nil, err
	}

	return NewFileStore(tmpdir)
}

func NewFileStore(dir string) (*FileStore, error) {
	store, err := local.NewStore(dir)
	if err != nil {
		return nil, err
	}
	return &FileStore{
		store: store,
	}, nil
}

func (s *FileStore) Write(ctx context.Context, ref string, descriptor ocispec.Descriptor, blob []byte) error {
	writer, err := s.store.Writer(ctx, content.WithRef(ref))
	if err != nil {
		return err
	}
	_, err = writer.Write(blob)
	if err != nil {
		return err
	}
	if err := writer.Commit(ctx, int64(len(blob)), digest.FromBytes(blob)); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

func (s *FileStore) Push(ctx context.Context, resolver remotes.Resolver, ref string, image *image.Descriptor) (*digest.Digest, error) {
	pusher, err := resolver.Pusher(ctx, ref)
	if err != nil {
		return nil, err
	}

	if err := remotes.PushContent(ctx, pusher, image.Manifest, s.store, nil, nil); err != nil {
		return nil, err
	}
	return &image.Manifest.Digest, nil
}

