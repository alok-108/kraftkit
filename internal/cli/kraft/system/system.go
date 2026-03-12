// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2025, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package system

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"kraftkit.sh/cmdfactory"

	"kraftkit.sh/internal/cli/kraft/system/list"
	"kraftkit.sh/internal/cli/kraft/system/set"
	"kraftkit.sh/internal/cli/kraft/system/unset"
)

type System struct{}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&System{}, cobra.Command{
		Short:   "Manage KraftKit and host system",
		Use:     "system SUBCOMMAND",
		Aliases: []string{"sys", "self"},
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup:  "misc",
			cmdfactory.AnnotationHelpHidden: "true",
		},
	})
	if err != nil {
		panic(err)
	}

	cmd.AddCommand(set.NewCmd())
	cmd.AddCommand(list.NewCmd())
	cmd.AddCommand(unset.NewCmd())

	return cmd
}

func (opts *System) Run(ctx context.Context, args []string) error {
	return pflag.ErrHelp
}
