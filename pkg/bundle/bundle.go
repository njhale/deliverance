package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	orascontent "github.com/deislabs/oras/pkg/content"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
)

type Blob struct {
	MediaType string
	Content   []byte
}

func PushDir(ctx context.Context, resolver remotes.Resolver, ref, dir string) error {
	memoryStore := orascontent.NewMemoryStore()
	pushContents := []ocispec.Descriptor{}

	blob, hash, err := DirLayerBlobReader(dir)
	if err != nil {
		return err
	}

	pushContents = append(pushContents, memoryStore.Add("manifests", string(blob.MediaType), blob.Content))

	fmt.Printf("Pushing to %s...\n", ref)
	desc, err := PushManifest(ctx, resolver, ref, memoryStore, pushContents, []digest.Digest{hash})
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
