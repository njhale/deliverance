package filestore

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/ecordell/bndlr/pkg/image"
	"github.com/ecordell/bndlr/pkg/image/layer"
	"github.com/ecordell/bndlr/pkg/registry/store"
)

type FileStore struct {
	store content.Store
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

func (s *FileStore) Pull(ctx context.Context, resolver remotes.Resolver, ref string) (*image.Descriptor, error) {
	_, root, err := resolver.Resolve(ctx, ref)
	if err != nil {
		return nil, err
	}

	fetcher, err := resolver.Fetcher(ctx, ref)
	if err != nil {
		return nil, err
	}

	descs, err := fetch(ctx, fetcher, root, s.store)
	if err != nil {
		return nil, err
	}
	img := image.NewDescriptor(descs)

	return img, err
}

func (s *FileStore) Unpack(ctx context.Context, img *image.Descriptor, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	for _, l := range img.Layers {
		r, err := s.store.ReaderAt(ctx, l)
		if err != nil {
			return err
		}

		blob := make([]byte, r.Size())
		if _, err = r.ReadAt(blob, 0); err != nil {
			return err
		}

		err = layer.LayerToDirectory(dir, &layer.Layer{
			Blob:   blob,
			Digest: l.Digest,
		})
		if err != nil {
			return err
		}

		r.Close()
	}

	return nil
}

func fetch(ctx context.Context, fetcher remotes.Fetcher, root ocispec.Descriptor, store content.Store) (descs []ocispec.Descriptor, err error) {
	var lock sync.Mutex
	visitor := images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		lock.Lock()
		defer lock.Unlock()
		fmt.Printf("desc: %v\n", desc)
		descs = append(descs, desc)
		return nil, nil
	})

	handler := images.Handlers(
		visitor,
		remotes.FetchHandler(store, fetcher),
		images.ChildrenHandler(store),
	)
	err = images.Dispatch(ctx, handler, nil, root)

	return
}
