package manifest

import (
	"encoding/json"

	"github.com/containerd/containerd/images"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/ecordell/bndlr/pkg/image/layer"
)

// A ManifestDescriptor can generate manifests for a given target
type ManifestDescriptor interface {
	MakeManifest(config ocispec.Descriptor, layers []ocispec.Descriptor) ([]byte, ocispec.Descriptor, error)
}

type ManifestDescriptorFunc func(config ocispec.Descriptor, layers []ocispec.Descriptor) ([]byte, ocispec.Descriptor, error)

func (f ManifestDescriptorFunc) MakeManifest(config ocispec.Descriptor, layers []ocispec.Descriptor) ([]byte, ocispec.Descriptor, error) {
	return f(config, layers)
}

// A ConfigDescriptor can create an image config given a list of digests
type ConfigDescriptor interface {
	MakeConfig(digests []digest.Digest) ([]byte, ocispec.Descriptor, error)
}

type ConfigDescriptorFunc func(digests []digest.Digest) ([]byte, ocispec.Descriptor, error)

func (f ConfigDescriptorFunc) MakeConfig(digests []digest.Digest) ([]byte, ocispec.Descriptor, error) {
	return f(digests)
}

// A LayerDescriptor can create an image config given a list of digests
type LayerDescriptor interface {
	MakeDescriptor(l layer.Layer) ([]byte, ocispec.Descriptor, error)
}

type LayerDescriptorFunc func(l layer.Layer) ([]byte, ocispec.Descriptor, error)

func (f LayerDescriptorFunc) MakeDescriptor(l layer.Layer) ([]byte, ocispec.Descriptor, error) {
	return f(l)
}

var _ ManifestDescriptorFunc = NewV22Manifest
var _ ConfigDescriptorFunc = NewMinimalV22Config
var _ LayerDescriptorFunc = NewLayerDescriptor

// NewV22Manifest returns a valid v2-2 manifest given a config and layers
func NewV22Manifest(config ocispec.Descriptor, layers []ocispec.Descriptor) ([]byte, ocispec.Descriptor, error) {
	manifest := struct {
		SchemaVersion int                  `json:"schemaVersion"`
		MediaType     string               `json:"mediaType"`
		Config        ocispec.Descriptor   `json:"config"`
		Layers        []ocispec.Descriptor `json:"layers"`
	}{
		SchemaVersion: 2,
		MediaType:     images.MediaTypeDockerSchema2Manifest,
		Config:        config,
		Layers:        layers,
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, ocispec.Descriptor{}, err
	}
	manifestDescriptor := ocispec.Descriptor{
		MediaType: images.MediaTypeDockerSchema2Manifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	return manifestBytes, manifestDescriptor, nil
}

// NewV22Config returns a minimal v2-2 config manifest. `digests` contain the digests of the uncompressed layers.
func NewMinimalV22Config(digests []digest.Digest) ([]byte, ocispec.Descriptor, error) {
	// Config Descriptor describes the content
	// Includes DiffIDs for docker compatibility
	imgconfig := ocispec.Image{
		// Not required
		OS: "linux",
		// Not required
		Architecture: "amd64",
		// Required by docker/distribution registries
		RootFS: ocispec.RootFS{
			Type:    "layers",
			DiffIDs: digests,
		},
		// Required by quay
		History: []ocispec.History{
			{
				CreatedBy: "bndlr",
			},
		},
	}

	configBytes, err := json.Marshal(imgconfig)
	if err != nil {
		return nil, ocispec.Descriptor{}, err
	}
	return configBytes, ocispec.Descriptor{
		MediaType: images.MediaTypeDockerSchema2Config,
		Digest:    digest.FromBytes(configBytes),
		Size:      int64(len(configBytes)),
	}, nil
}

func NewLayerDescriptor(l layer.Layer) ([]byte, ocispec.Descriptor, error) {
	desc := ocispec.Descriptor{
		MediaType:   l.MediaType,
		Digest:      digest.FromBytes(l.Blob),
		Size:        int64(len(l.Blob)),
	}
	layerBytes, err := json.Marshal(desc)
	if err != nil {
		return nil, ocispec.Descriptor{}, err
	}
	return layerBytes, desc, nil
}
