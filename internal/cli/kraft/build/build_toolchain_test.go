// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package build

import (
	"reflect"
	"testing"
)

func TestMergeToolchain(t *testing.T) {
	for _, tc := range []struct {
		name   string
		global map[string]string
		flags  []string
		expect map[string]string
	}{
		{
			name:   "flags only",
			global: nil,
			flags:  []string{"CC=clang", "LD=ld.lld"},
			expect: map[string]string{"CC": "clang", "LD": "ld.lld"},
		},
		{
			name:   "global only",
			global: map[string]string{"CC": "gcc", "UK_CFLAGS": "-O2"},
			flags:  nil,
			expect: map[string]string{"CC": "gcc", "UK_CFLAGS": "-O2"},
		},
		{
			name:   "flags and global disjoint keys",
			global: map[string]string{"UK_CFLAGS": "-O2"},
			flags:  []string{"CC=clang"},
			expect: map[string]string{"CC": "clang", "UK_CFLAGS": "-O2"},
		},
		{
			name:   "flag overrides global on same key",
			global: map[string]string{"CC": "gcc"},
			flags:  []string{"CC=clang"},
			expect: map[string]string{"CC": "clang"},
		},
		{
			name:   "malformed flag without equals is ignored",
			global: nil,
			flags:  []string{"BADENTRY", "CC=clang"},
			expect: map[string]string{"CC": "clang"},
		},
		{
			name:   "both empty returns empty map",
			global: nil,
			flags:  nil,
			expect: map[string]string{},
		},
		{
			name:   "value containing equals sign is preserved",
			global: nil,
			flags:  []string{"UK_CFLAGS=-O2 -DFOO=1"},
			expect: map[string]string{"UK_CFLAGS": "-O2 -DFOO=1"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeToolchain(tc.global, tc.flags)
			if !reflect.DeepEqual(got, tc.expect) {
				t.Errorf("mergeToolchain() = %v, want %v", got, tc.expect)
			}
		})
	}
}
