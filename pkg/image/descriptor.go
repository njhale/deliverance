package image

import (
	"github.com/opencontainers/image-spec/specs-go/v1"
)

type Descriptor struct {
	Manifest v1.Descriptor
	Config   v1.Descriptor
	Layers   []v1.Descriptor
}
