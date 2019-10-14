package layer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
)

// Layer represents a single layer: a compressed blob and a digest of the uncompressed blob
type Layer struct {
	Blob      []byte
	Digest    digest.Digest
	MediaType string
	Name      string
}

type Layers []Layer

func (l Layers) Digests() (digests []digest.Digest) {
	for _, d := range l {
		digests = append(digests, d.Digest)
	}
	return
}

type LayerOption func(config *Layer)

// apply sequentially applies the given options to the layer.
func (c *Layer) apply(options []LayerOption) *Layer {
	for _, option := range options {
		option(c)
	}
	return c
}

func WithMediaType(mediaType string) LayerOption {
	return func(layer *Layer) {
		layer.MediaType = mediaType
	}
}

func WithName(name string) LayerOption {
	return func(layer *Layer) {
		layer.Name = name
	}
}

// LayerFromDirectory builds a single tgz image layer from a directory of files
// returns the gzip data in a byte buffer and the digest of the uncompressed files
func LayerFromDirectory(directory string, opts ...LayerOption) (*Layer, error) {
	if _, err := os.Stat(directory); err != nil {
		return nil, err
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
		return nil, err
	}

	// close writer here to get the correct hash - defer will not work
	if err := writer.Close(); err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(&buf)
	if err != nil {
		return nil, err
	}
	l := &Layer{
		Blob: b,
		Digest: digest.NewDigestFromBytes(digest.SHA256, hash.Sum(nil)),
	}
	return l.apply(opts), nil
}
