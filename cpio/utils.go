// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2025, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package cpio

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"

	"github.com/anchore/stereoscope"
)

// IsCpioFile checks if the given file is a cpio archive.
func IsCpioFile(path string) bool {
	fi, err := os.Open(path)
	if err != nil {
		return false
	}
	defer fi.Close()

	reader := NewReader(fi)

	_, _, err = reader.Next()
	if err != nil {
		return false
	}

	return true
}

// IsTarFile checks if the given file is a tar archive.
func IsTarFile(path string) bool {
	fi, err := os.Open(path)
	if err != nil {
		return false
	}
	defer fi.Close()

	tarReader := tar.NewReader(fi)
	_, err = tarReader.Next()
	if err != nil && err != io.EOF {
		return false
	}

	return true
}

// IsTarGzFile checks if the given file is a gzipped tar archive.
func IsTarGzFile(path string) bool {
	fi, err := os.Open(path)
	if err != nil {
		return false
	}
	defer fi.Close()

	gzr, err := gzip.NewReader(fi)
	if err != nil {
		return false
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)
	_, err = tarReader.Next()
	if err != nil && err != io.EOF {
		return false
	}

	return true
}

// IsDirectory checks if the given path is a directory.
func IsDirectory(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// IsOciArchive checks if the given path is an OCI archive.
func IsOciArchive(path string) bool {
	image, err := stereoscope.GetImage(context.TODO(), path)
	if err != nil {
		return false
	}
	defer image.Cleanup()

	return image != nil && image.SquashedTree() != nil
}
