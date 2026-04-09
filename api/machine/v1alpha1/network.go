// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2024, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package v1alpha1

import (
	"fmt"
	"strings"
)

// MachineNetwork represents a parsed network configuration from a CLI string.
type MachineNetwork struct {
	Name     string
	CIDR     string
	Gateway  string
	DNS0     string
	DNS1     string
	Hostname string
	Domain   string
}

// ParseNetwork parses a colon-separated network string into a MachineNetwork.
// The format is: name[:cidr[:gateway[:dns0[:dns1[:hostname[:domain]]]]]].
// Only the name is required.
func ParseNetwork(s string) (*MachineNetwork, error) {
	if s == "" {
		return nil, fmt.Errorf("network string cannot be empty")
	}

	parts := strings.SplitN(s, ":", 7)
	if parts[0] == "" {
		return nil, fmt.Errorf("network name cannot be empty")
	}

	mn := &MachineNetwork{Name: parts[0]}

	// Assign optional fields if they exist.
	if len(parts) > 1 {
		mn.CIDR = parts[1]
	}
	if len(parts) > 2 {
		mn.Gateway = parts[2]
	}
	if len(parts) > 3 {
		mn.DNS0 = parts[3]
	}
	if len(parts) > 4 {
		mn.DNS1 = parts[4]
	}
	if len(parts) > 5 {
		mn.Hostname = parts[5]
	}
	if len(parts) > 6 {
		mn.Domain = parts[6]
	}

	return mn, nil
}

// String implements fmt.Stringer and returns the colon-separated
// representation of the MachineNetwork.
func (mn *MachineNetwork) String() string {
	fields := []string{
		mn.Name,
		mn.CIDR,
		mn.Gateway,
		mn.DNS0,
		mn.DNS1,
		mn.Hostname,
		mn.Domain,
	}

	// Trim trailing empty strings.
	for i := len(fields) - 1; i >= 0; i-- {
		if fields[i] != "" {
			fields = fields[:i+1]
			break
		}
	}

	return strings.Join(fields, ":")
}
