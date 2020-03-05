package filestore

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ecordell/bndlr/pkg/registry"
)

func TestPull(t *testing.T) {
	fs, err := NewFileStore("test-fs")
	require.NoError(t, err)

	ref := "quay.io/olmtest/kiali:1.2.4"
	resolver := registry.NewResolver("", "")
	ctx := context.Background()
	img, err := fs.Pull(ctx, resolver, ref)
	require.NoError(t, err)
	fmt.Printf("img: %v", img)

	dir := "test-bundle"
	err = fs.Unpack(ctx, img, dir)
	require.NoError(t, err)
}
