// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package initrd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type file struct {
	opts InitrdOptions
	path string
}

// NewFromFile accepts an input file which already represents a CPIO archive and
// is provided as a mechanism for satisfying the Initrd interface.
func NewFromFile(_ context.Context, path string, opts ...InitrdOption) (Initrd, error) {
	initrd := file{
		opts: InitrdOptions{},
		path: path,
	}

	for _, opt := range opts {
		if err := opt(&initrd.opts); err != nil {
			return nil, err
		}
	}

	if !filepath.IsAbs(initrd.path) {
		initrd.path = filepath.Join(initrd.opts.workdir, initrd.path)
	}

	stat, err := os.Stat(initrd.path)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("path %s is a directory, not a file", initrd.path)
	}

	absDest, err := filepath.Abs(filepath.Clean(initrd.opts.output))
	if err != nil {
		return nil, fmt.Errorf("getting absolute path of destination: %w", err)
	}

	if absDest == stat.Name() {
		return nil, fmt.Errorf("CPIO archive path is the same as the source path, this is not allowed as it creates corrupted archives")
	}

	return &initrd, nil
}

// Build implements Initrd.
func (initrd *file) Name() string {
	return "file"
}

// Build implements Initrd.
func (initrd *file) Build(ctx context.Context) (string, error) {
	switch initrd.opts.fsType {
	case FsTypeErofs:
		isCpio, err := isCpioFile(initrd.path)
		if err != nil {
			return "", fmt.Errorf("could not determine if file is a CPIO archive: %w", err)
		}

		isErofs, err := isErofsFile(initrd.path)
		if err != nil {
			return "", fmt.Errorf("could not determine if file is an EROFS archive: %w", err)
		}

		if isCpio {
			return "", fmt.Errorf("CPIO-to-EROFS conversion currently not supported. Use 'bsdcpio' or 'cpio' to unpack it first to a directory")
		}

		if !isErofs {
			return "", fmt.Errorf("file %s is not a valid EROFS archive", initrd.path)
		}
	case FsTypeCpio:
		isCpio, err := isCpioFile(initrd.path)
		if err != nil {
			return "", fmt.Errorf("could not determine if file is a CPIO archive: %w", err)
		}

		isErofs, err := isErofsFile(initrd.path)
		if err != nil {
			return "", fmt.Errorf("could not determine if file is an EROFS archive: %w", err)
		}

		if isErofs {
			return "", fmt.Errorf("EROFS-to-CPIO conversion currently not supported. Use 'fsck.erofs' to unpack it first to a directory")
		}

		if !isCpio {
			return "", fmt.Errorf("file %s is not a valid CPIO archive", initrd.path)
		}
	default:
		return "", fmt.Errorf("unknown filesystem type %s for file %s", initrd.opts.fsType, initrd.path)
	}

	return initrd.path, nil
}

// Options implements Initrd.
func (initrd *file) Options() InitrdOptions {
	return initrd.opts
}

// Env implements Initrd.
func (initrd *file) Env() []string {
	return nil
}

// Args implements Initrd.
func (initrd *file) Args() []string {
	return nil
}
