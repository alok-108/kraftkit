// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2025, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package erofs

type ErofsCreateOptions struct {
	allRoot bool
}

type ErofsCreateOption func(*ErofsCreateOptions) error

// WithAllRoot toggles whether all files permissions should be set to root:root
// instead of the original file permissions.
func WithAllRoot(allRoot bool) ErofsCreateOption {
	return func(eo *ErofsCreateOptions) error {
		eo.allRoot = allRoot
		return nil
	}
}
