package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	orascontent "github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
)

// PushDir creates a v2-2 image with a single layer that contains the contents
// of a single directory
func PushDir(ctx context.Context, resolver remotes.Resolver, ref, dir string) error {
	memoryStore := orascontent.NewMemoryStore()
	layers := []ocispec.Descriptor{}

	blob, hash, err := BuildLayer(dir)
	if err != nil {
		return err
	}

	layers = append(layers, memoryStore.Add("manifests", images.MediaTypeDockerSchema2LayerGzip, blob))

	now := time.Now()
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
			DiffIDs: []digest.Digest{hash},
		},
		// Required by quay
		History: []ocispec.History{
			{
				Created: &now,
			},
		},
	}
	configBytes, err := json.Marshal(imgconfig)
	if err != nil {
		return err
	}
	config := ocispec.Descriptor{
		MediaType: images.MediaTypeDockerSchema2Config,
		Digest:    digest.FromBytes(configBytes),
		Size:      int64(len(configBytes)),
	}
	memoryStore.Set(config, configBytes)

	// Manifest describes the layers
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
		return err
	}
	manifestDescriptor := ocispec.Descriptor{
		MediaType: images.MediaTypeDockerSchema2Manifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	memoryStore.Set(manifestDescriptor, manifestBytes)

	fmt.Printf("Pushing to %s...\n", ref)

	desc, err := oras.Push(ctx, resolver, ref, memoryStore, layers, oras.WithConfig(config), oras.WithManifest(manifestDescriptor))
	if err != nil {
		return err
	}

	fmt.Printf("Pushed  with digest %s\n", desc.Digest)
	return nil
}

// BuildLayer builds a single tgz image layer from a directory of files
// returns the gzip data in a byte buffer and the digest of the uncompressed files
func BuildLayer(directory string) ([]byte, digest.Digest, error) {
	if _, err := os.Stat(directory); err != nil {
		return nil, "", err
	}

	// set up our layer pipeline
	//
	//                  -> gz -> byte buffer
	//                /
	// files -> tar -
	//                \
	//                  -> sha256 -> digest
	//

	// the byte buffer contains compressed layer data,
	// and the hash is the digest of the uncompressed layer
	// data, which docker requires (oci does not)

	// output writers
	hash := sha256.New()
	var buf bytes.Buffer

	// from gzip to buffer
	gzipWriter := gzip.NewWriter(&buf)

	// from files to hash/gz
	hashAndGzWriter := io.MultiWriter(hash, gzipWriter)
	writer := tar.NewWriter(hashAndGzWriter)

	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		err = writer.WriteHeader(header)
		if err != nil {
			return err
		}

		// if it's a directory, just write the header and continue
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			if err := file.Close(); err != nil {
				logrus.Warnf("error closing file: %s", err.Error())
			}
		}()

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		return err
	}); err != nil {
		return nil, "", err
	}

	// close writer here to get the correct hash - defer will not work
	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, "", err
	}

	b, err := ioutil.ReadAll(&buf)
	if err != nil {
		return nil, "", err
	}

	return b, digest.NewDigestFromBytes(digest.SHA256, hash.Sum(nil)), nil
}
