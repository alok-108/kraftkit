// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2024, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kraftkit.sh/unikraft/app"
)

func TestNewProjectFromOptionsCustomKraftfile(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "kraftkit-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	customName := "CustomKraftfile.yaml"
	customPath := filepath.Join(tmpDir, customName)
	content := []byte(`
spec: v0.6
name: test-project
unikraft: v0.16.3
`)

	err = os.WriteFile(customPath, content, 0o644)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("recognize custom kraftfile name", func(t *testing.T) {
		project, err := app.NewProjectFromOptions(ctx,
			app.WithProjectWorkdir(tmpDir),
			app.WithProjectKraftfile(customPath),
		)
		require.NoError(t, err)
		require.NotNil(t, project)
		assert.Equal(t, "test-project", project.Name())
	})

	t.Run("fail when no kraftfile exists in workdir", func(t *testing.T) {
		// New directory with no kraftfile
		emptyDir, err := os.MkdirTemp("", "kraftkit-empty-*")
		require.NoError(t, err)
		defer os.RemoveAll(emptyDir)

		project, err := app.NewProjectFromOptions(ctx,
			app.WithProjectWorkdir(emptyDir),
			app.WithProjectDefaultKraftfiles(),
		)
		assert.Error(t, err)
		assert.Nil(t, project)
		assert.Contains(t, err.Error(), "no Kraftfile specified")
	})
}
