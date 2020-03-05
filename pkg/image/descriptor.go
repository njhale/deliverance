package image

import (
	"github.com/containerd/containerd/images"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type Descriptor struct {
	Manifest ocispec.Descriptor
	Index    ocispec.Descriptor
	Config   ocispec.Descriptor
	Layers   []ocispec.Descriptor
	Unknown  []ocispec.Descriptor
}

func NewDescriptor(descs []ocispec.Descriptor) *Descriptor {
	if len(descs) == 0 {
		return nil
	}

	img := &Descriptor{}
	for _, desc := range descs {
		switch desc.MediaType {
		case ocispec.MediaTypeImageIndex, images.MediaTypeDockerSchema2ManifestList:
			img.Index = desc
		case ocispec.MediaTypeImageManifest, images.MediaTypeDockerSchema2Manifest:
			img.Manifest = desc
		case ocispec.MediaTypeImageConfig, images.MediaTypeDockerSchema2Config:
			img.Config = desc
		case ocispec.MediaTypeImageLayer, ocispec.MediaTypeImageLayerGzip,
			images.MediaTypeDockerSchema2Layer, images.MediaTypeDockerSchema2LayerGzip,
			images.MediaTypeDockerSchema2LayerForeign, images.MediaTypeDockerSchema2LayerForeignGzip:
			img.Layers = append(img.Layers, desc)
		default:
			// Catch all mediatypes we don't know about as unknown
			img.Unknown = append(img.Unknown, desc)
		}
	}

	return img
}
