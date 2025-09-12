// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2025, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package erofs

import (
	"os"
)

// IsErofsFile checks if the given file is an EROFS filesystem.
func IsErofsFile(initrd string) bool {
	fi, err := os.Open(initrd)
	if err != nil {
		return false
	}
	defer fi.Close()

	_, err = Open(fi)
	return err == nil
}
