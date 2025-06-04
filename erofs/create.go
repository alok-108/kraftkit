// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2025, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package erofs

import (
	"context"
	"fmt"
	"io"
	"os"

	archive "kraftkit.sh/cpio"
)

type createOptions struct{}

func CreateFS(ctx context.Context, output string, source string, opts ...ErofsCreateOption) error {
	c := &createOptions{}

	// Open writer for the output file
	writer, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("could not open output file: %w", err)
	}
	defer writer.Close()

	switch {
	case archive.IsOciArchive(source):
		if err := c.CreateFSFromOCIImage(ctx, writer, source, opts...); err != nil {
			return fmt.Errorf("could not create EroFS archive from OCI image: %w", err)
		}
	case archive.IsTarFile(source),
		archive.IsTarGzFile(source):
		if err := c.CreateFSFromTarFile(ctx, writer, source, opts...); err != nil {
			return fmt.Errorf("could not create EroFS archive from tar file: %w", err)
		}
	case archive.IsDirectory(source):
		if err := c.CreateFSFromDirectory(ctx, writer, source, opts...); err != nil {
			return fmt.Errorf("could not create EroFS archive from directory: %w", err)
		}
	case IsErofsFile(source):
		if err := c.CreateFSFromErofs(ctx, writer, source, opts...); err != nil {
			return fmt.Errorf("could not create EroFS archive from CPIO file: %w", err)
		}
	case archive.IsCpioFile(source):
		return fmt.Errorf("CPIO files are not currently supported as source for EroFS archives. Contributions welcome!")
	default:
		return fmt.Errorf("unsupported source type: %s", source)
	}

	return nil
}

// CreateFSFromOCIImage creates an EroFS filesystem from an OCI image.
func (c *createOptions) CreateFSFromOCIImage(ctx context.Context, writer *os.File, source string, opts ...ErofsCreateOption) error {
	source, err := unpackOCIImageToDirectory(ctx, source)
	if err != nil {
		return fmt.Errorf("could not unpack OCI file: %w", err)
	}
	defer os.RemoveAll(source)

	if err := c.CreateFSFromDirectory(ctx, writer, source, opts...); err != nil {
		return fmt.Errorf("could not create EroFS archive from directory: %w", err)
	}

	return nil
}

// CreateFSFromDirectory creates an EroFS filesystem from a directory.
func (c *createOptions) CreateFSFromDirectory(ctx context.Context, writer *os.File, source string, opts ...ErofsCreateOption) error {
	return Create(io.WriterAt(writer), os.DirFS(source), opts...)
}

// CreateFSFromTarFile creates an EroFS filesystem from a tar file.
func (c *createOptions) CreateFSFromTarFile(ctx context.Context, writer *os.File, source string, opts ...ErofsCreateOption) error {
	source, err := unpackTarFileToDirectory(ctx, source)
	if err != nil {
		return fmt.Errorf("could not unpack tar file: %w", err)
	}
	defer os.RemoveAll(source)

	if err := c.CreateFSFromDirectory(ctx, writer, source, opts...); err != nil {
		return fmt.Errorf("could not create EroFS archive from directory: %w", err)
	}

	return nil
}

// CreateFSFromErofs creates an EroFS filesystem from an existing EroFS file.
func (c *createOptions) CreateFSFromErofs(ctx context.Context, writer *os.File, source string, opts ...ErofsCreateOption) error {
	// Open and copy all contents from 'source' to the writer
	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("could not open EroFS source file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(writer, f)
	if err != nil {
		return fmt.Errorf("could not copy EroFS data: %w", err)
	}

	return nil
}
