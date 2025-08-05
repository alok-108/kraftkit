// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package initrd

type InitrdOptions struct {
	compress bool
	output   string
	cacheDir string
	arch     string
	workdir  string
}

// Whether the resulting CPIO archive file should be compressed.
func (opts InitrdOptions) Compress() bool {
	return opts.compress
}

// The output location of the resulting CPIO archive file.
func (opts InitrdOptions) Output() string {
	return opts.output
}

// The cache directory used during the serialization of the initramfs.
func (opts InitrdOptions) CacheDir() string {
	return opts.cacheDir
}

// The architecture of the file contents of binaries in the initramfs.
func (opts InitrdOptions) Architecture() string {
	return opts.arch
}

// The working directory of the initramfs builder.
func (opts InitrdOptions) Workdir() string {
	return opts.workdir
}

type InitrdOption func(*InitrdOptions) error

// WithCompression sets the compression of the resulting CPIO archive file.
func WithCompression(compress bool) InitrdOption {
	return func(opts *InitrdOptions) error {
		opts.compress = compress
		return nil
	}
}

// WithOutput sets the location of the output location of the resulting CPIO
// archive file.
func WithOutput(output string) InitrdOption {
	return func(opts *InitrdOptions) error {
		opts.output = output
		return nil
	}
}

// WithCacheDir sets the path of an internal location that's used during the
// serialization of the initramfs as a mechanism for storing temporary files
// used as cache.
func WithCacheDir(dir string) InitrdOption {
	return func(opts *InitrdOptions) error {
		opts.cacheDir = dir
		return nil
	}
}

// WithArchitecture sets the architecture of the file contents of binaries in
// the initramfs.  Files may not always be architecture specific, this option
// simply indicates the target architecture if any binaries are compiled by the
// implementing initrd builder.
func WithArchitecture(arch string) InitrdOption {
	return func(opts *InitrdOptions) error {
		opts.arch = arch
		return nil
	}
}

// WithWorkdir sets the working directory of the initramfs builder.  This is
// used as a mechanism for storing temporary files and directories during the
// serialization of the initramfs.
func WithWorkdir(dir string) InitrdOption {
	return func(opts *InitrdOptions) error {
		opts.workdir = dir
		return nil
	}
}
