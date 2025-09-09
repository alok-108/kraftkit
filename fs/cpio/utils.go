// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2024, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package cpio

import "os"

// IsCpioFile checks if the given file is a cpio archive.
func IsCpioFile(path string) bool {
	fi, err := os.Open(path)
	if err != nil {
		return false
	}
	defer fi.Close()

	reader := NewReader(fi)

	_, _, err = reader.Next()
	return err == nil
}
