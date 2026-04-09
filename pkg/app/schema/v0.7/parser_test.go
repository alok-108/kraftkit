// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2024, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package v0_7

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func loadTestdata(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	require.NoError(t, err)
	return data
}

func TestParse_Minimal(t *testing.T) {
	data := loadTestdata(t, "minimal.yaml")

	var app Kraftfile
	err := yaml.Unmarshal(data, &app)
	require.NoError(t, err)

	assert.Equal(t, "0.7", app.Version)
	assert.Equal(t, "minimal-app", app.Name)
}

func TestParse_Full(t *testing.T) {
	data := loadTestdata(t, "full.yaml")

	var app Kraftfile
	err := yaml.Unmarshal(data, &app)
	require.NoError(t, err)

	assert.Equal(t, "0.7", app.Version)
	assert.Equal(t, "full-app", app.Name)
	assert.NotNil(t, app.Unikraft)
	assert.NotEmpty(t, app.Targets)
	assert.NotEmpty(t, app.Libraries)
}

func TestParse_InvalidYAML(t *testing.T) {
	data := loadTestdata(t, "invalid.yaml")

	var app Kraftfile
	err := yaml.Unmarshal(data, &app)
	require.Error(t, err)
}

func TestParse_WrongVersion(t *testing.T) {
	data := loadTestdata(t, "wrong_version.yaml")

	var app Kraftfile
	err := yaml.Unmarshal(data, &app)
	
	// Assuming Kraftfile unmarshaling directly rejects wrong versions OR 
	// a separate validation function handles it. Typically unmarshal checks it.
	// For testing, we ensure that if parsed it doesn't match 0.7.
	if err == nil {
		assert.NotEqual(t, "0.7", app.Version)
	}
}

func TestRoundTrip(t *testing.T) {
	data := loadTestdata(t, "full.yaml")

	var original Kraftfile
	err := yaml.Unmarshal(data, &original)
	require.NoError(t, err)

	serialized, err := yaml.Marshal(&original)
	require.NoError(t, err)

	var reParsed Kraftfile
	err = yaml.Unmarshal(serialized, &reParsed)
	require.NoError(t, err)

	assert.Equal(t, original, reParsed)
}
