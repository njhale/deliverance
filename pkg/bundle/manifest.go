package bundle

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	orascontent "github.com/deislabs/oras/pkg/content"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)



// Manifest replaces ocispec.Manifest, which does not contain MediaType
// docker complains about missing mediatype if you give non-oci mediatypes
type Manifest struct {
	specs.Versioned
	images.Image

	MediaType string `json:"mediaType"`

	// Config references a configuration object for a container, by digest.
	// The referenced configuration object is a JSON blob that the runtime uses to set up the container.
	Config ocispec.Descriptor `json:"config"`

	// Layers is an indexed list of layers referenced by the manifest.
	Layers []ocispec.Descriptor `json:"layers"`

	// Annotations contains arbitrary metadata for the image manifest.
	Annotations map[string]string `json:"annotations,omitempty"`
}

func PushManifest(ctx context.Context, resolver remotes.Resolver, ref string, provider content.Provider, descriptors []ocispec.Descriptor, digests []digest.Digest) (ocispec.Descriptor, error) {
	pusher, err := resolver.Pusher(ctx, ref)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	desc, provider, err := pack(provider, descriptors, digests)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	if err := remotes.PushContent(ctx, pusher, desc, provider, nil, pushStatusTrack()); err != nil {
		return ocispec.Descriptor{}, err
	}
	return desc, nil
}

func pack(provider content.Provider, descriptors []ocispec.Descriptor, digests []digest.Digest) (ocispec.Descriptor, content.Provider, error) {
	store := newHybridStoreFromProvider(provider)

	// Config
	imgconfig := ocispec.Image{
		RootFS: ocispec.RootFS{
			Type:    "layers",
			DiffIDs: digests,
		},
	}
	configBytes, err := json.Marshal(imgconfig)
	if err != nil {
		return ocispec.Descriptor{}, nil, err
	}
	config := ocispec.Descriptor{
		MediaType: images.MediaTypeDockerSchema2Config,
		Digest:    digest.FromBytes(configBytes),
		Size:      int64(len(configBytes)),
	}
	store.Set(config, configBytes)

	// Manifest
	manifest := Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2, // historical value. does not pertain to OCI or docker version
		},
		MediaType: images.MediaTypeDockerSchema2Manifest,
		Config:    config,
		Layers:    descriptors,
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return ocispec.Descriptor{}, nil, err
	}

	manifestDescriptor := ocispec.Descriptor{
		MediaType: images.MediaTypeDockerSchema2Manifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	store.Set(manifestDescriptor, manifestBytes)

	return manifestDescriptor, store, nil
}

func pushStatusTrack() images.Handler {
	var printLock sync.Mutex
	return images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		if name, ok := orascontent.ResolveName(desc); ok {
			printLock.Lock()
			defer printLock.Unlock()
			fmt.Println("Uploading", desc.Digest.Encoded()[:12], name)
		}
		return nil, nil
	})
}



// ensure interface
var (
	_ content.Provider = &hybridStore{}
	_ content.Ingester = &hybridStore{}
)

type hybridStore struct {
	cache    *orascontent.Memorystore
	provider content.Provider
	ingester content.Ingester
}

func newHybridStoreFromProvider(provider content.Provider) *hybridStore {
	return &hybridStore{
		cache:    orascontent.NewMemoryStore(),
		provider: provider,
	}
}

func (s *hybridStore) Set(desc ocispec.Descriptor, content []byte) {
	s.cache.Set(desc, content)
}

// ReaderAt provides contents
func (s *hybridStore) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	readerAt, err := s.cache.ReaderAt(ctx, desc)
	if err == nil {
		return readerAt, nil
	}
	if s.provider != nil {
		return s.provider.ReaderAt(ctx, desc)
	}
	return nil, err
}

// Writer begins or resumes the active writer identified by desc
func (s *hybridStore) Writer(ctx context.Context, opts ...content.WriterOpt) (content.Writer, error) {
	var wOpts content.WriterOpts
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil, err
		}
	}

	if s.ingester == nil {
		return s.cache.Writer(ctx, opts...)
	}
	return s.ingester.Writer(ctx, opts...)
}
