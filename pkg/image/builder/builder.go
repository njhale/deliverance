package builder

import (
	"context"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/ecordell/bndlr/pkg/image"
	"github.com/ecordell/bndlr/pkg/image/layer"
	"github.com/ecordell/bndlr/pkg/image/manifest"
	"github.com/ecordell/bndlr/pkg/registry/store"
)

// Builder is used build manifests with particular configurations
// from a set of layers
type Builder struct {
	// manifestDescriptor knows how to build a manifest
	manifestDescriptor manifest.ManifestDescriptor

	// manifestDescriptor knows how to build a config descriptor
	configDescriptor manifest.ConfigDescriptor

	// manifestDescriptor knows how to build a single layer
	layerDescriptor manifest.LayerDescriptor
}


// NewMinimalV22ImageBuilder creates a v2-2 image with minimal metadata
func NewMinimalV22Builder() (*Builder, error) {
	return &Builder{
		manifestDescriptor: manifest.ManifestDescriptorFunc(manifest.NewV22Manifest),
		configDescriptor:   manifest.ConfigDescriptorFunc(manifest.NewMinimalV22Config),
		layerDescriptor:    manifest.LayerDescriptorFunc(manifest.NewLayerDescriptor),
	}, nil
}

// BuildImage builds a manifest and config from the configured layers and writes them into a store
func (c Builder) BuildImage(ctx context.Context, ref string, store store.Store, layers layer.Layers) (*image.Descriptor, error) {
	var layerDescs = make([]ocispec.Descriptor, 0)
	for _, l := range layers {
		if err := store.Write(ctx, ref, ocispec.Descriptor{}, l.Blob); err != nil {
			return nil, err
		}
		dbytes, d, err := c.layerDescriptor.MakeDescriptor(l)
		if err != nil {
			return nil, err
		}
		if err := store.Write(ctx, ref, d, dbytes); err != nil {
			return nil, err
		}
		layerDescs = append(layerDescs, d)
	}

	configBytes, config, err := c.configDescriptor.MakeConfig(layers.Digests())
	if err != nil {
		return nil, err
	}
	if err := store.Write(ctx, ref, config, configBytes); err != nil {
		return nil, err
	}

	manifestBytes, manifestDescriptor, err := c.manifestDescriptor.MakeManifest(config, layerDescs)
	if err != nil {
		return nil, err
	}
	if err := store.Write(ctx, ref, manifestDescriptor, manifestBytes); err != nil {
		return nil, err
	}
	return &image.Descriptor{
		Manifest: manifestDescriptor,
		Config:   config,
		Layers:   layerDescs,
	}, nil
}
