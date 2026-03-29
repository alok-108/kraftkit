// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package build

import (
	"bufio"
	"context"
	"fmt"
	"os"
	plainexec "os/exec"
	"strings"

	"kraftkit.sh/config"
	"kraftkit.sh/exec"
	"kraftkit.sh/log"
)

// mergeBuildEnv combines environment variables from multiple sources:
// 1. Config file (~/.kraft/config.yaml)
// 2. Profile (from config)
// 3. Environment file (--env-file)
// 4. CLI flags (--build-env)
func (opts *BuildOptions) mergeBuildEnv(ctx context.Context) (map[string]string, error) {
	env := make(map[string]string)

	// 1. Load from global config
	globalCfg := config.G[config.KraftKit](ctx)
	if globalCfg.Build.Env != nil {
		for k, v := range globalCfg.Build.Env {
			env[k] = v
		}
	}

	// 2. Load from profile
	if opts.Profile != "" {
		if globalCfg.Build.Profiles == nil {
			return nil, fmt.Errorf("no profiles defined in configuration")
		}
		profile, ok := globalCfg.Build.Profiles[opts.Profile]
		if !ok {
			return nil, fmt.Errorf("build profile not found: %s", opts.Profile)
		}
		for k, v := range profile {
			env[k] = v
		}
	}

	// 3. Load from env file
	if opts.EnvFile != "" {
		fileEnv, err := parseEnvFile(opts.EnvFile)
		if err != nil {
			return nil, fmt.Errorf("reading env file: %w", err)
		}
		for k, v := range fileEnv {
			env[k] = v
		}
	}

	// 4. Load from CLI flags
	for _, e := range opts.BuildEnv {
		if strings.ContainsRune(e, '=') {
			parts := strings.SplitN(e, "=", 2)
			env[parts[0]] = parts[1]
		} else {
			// If only key is provided, inherit from host
			env[e] = os.Getenv(e)
		}
	}

	if opts.DebugEnv {
		log.G(ctx).Info("Resulting build environment:")
		for k, v := range env {
			log.G(ctx).Infof("  %s=%s", k, v)
		}
	}

	// Basic validation
	if cc, ok := env["CC"]; ok {
		if _, err := os.Stat(cc); err != nil && !strings.ContainsRune(cc, '/') {
			// Check if it's in PATH
			if _, err := plainexec.LookPath(cc); err != nil {
				log.G(ctx).Warnf("Compiler %s not found in PATH", cc)
			}
		}
	}

	return env, nil
}

// BuildExecOptions returns a slice of exec.ExecOption for the merged environment
func (opts *BuildOptions) BuildExecOptions(ctx context.Context) ([]exec.ExecOption, error) {
	env, err := opts.mergeBuildEnv(ctx)
	if err != nil {
		return nil, err
	}

	var eopts []exec.ExecOption
	for k, v := range env {
		eopts = append(eopts, exec.WithEnvKey(k, v))
	}

	return eopts, nil
}

func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	env := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Remove quotes if present
			val = strings.Trim(val, `"'`)
			env[key] = val
		}
	}

	return env, scanner.Err()
}
