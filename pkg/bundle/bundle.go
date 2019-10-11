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

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	orascontent "github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
)

type Blob struct {
	MediaType string
	Content   []byte
}

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

func PushDir(ctx context.Context, resolver remotes.Resolver, ref, dir string) error {
	memoryStore := orascontent.NewMemoryStore()
	pushContents := []ocispec.Descriptor{}

	blob, hash, err := DirLayerBlobReader(dir)
	if err != nil {
		return err
	}

	pushContents = append(pushContents, memoryStore.Add("manifests", blob.MediaType, blob.Content))

	// Config
	imgconfig := ocispec.Image{
		RootFS: ocispec.RootFS{
			Type:    "layers",
			DiffIDs: []digest.Digest{hash},
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

	// Manifest
	manifest := Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2, // historical value. does not pertain to OCI or docker version
		},
		MediaType: images.MediaTypeDockerSchema2Manifest,
		Config:    config,
		Layers:    pushContents,
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

	desc, err := oras.Push(ctx,resolver, ref, memoryStore, pushContents, oras.WithConfig(config), oras.WithManifest(manifestDescriptor))
	if err != nil {
		return err
	}

	fmt.Printf("Pushed  with digest %s\n", desc.Digest)
	return nil
}


func DirLayerBlobReader(path string) (*Blob, digest.Digest, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, "", fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	b, d, err := BuildLayer(path)
	if err != nil {
		return nil, "", err
	}

	return &Blob{
		MediaType: images.MediaTypeDockerSchema2LayerGzip,
		Content: b,
	}, d, nil
}

func BuildLayer(directory string) ([]byte, digest.Digest, error) {

	// set up our layer pipline
	//                  -> gz -> byte buffer
	//                /
	// files -> tar -
	//                \
	//                  -> sha256 -> digest
	//

	// the byte buffer contains layer data,
	// and the hash is the digest of the uncompressed layer
	// data, which docker needs (oci does not)

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
