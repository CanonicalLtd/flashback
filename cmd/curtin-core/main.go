// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package main

import (
	"fmt"
	"os"

	"github.com/CanonicalLtd/curtin-core/execute"
	flags "github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewParser(&execute.Execution, flags.HelpFlag)
	_, err := parser.Parse()

	if err != nil {
		if e, ok := err.(*flags.Error); ok {
			if e.Type == flags.ErrHelp || e.Type == flags.ErrCommandRequired {
				parser.WriteHelp(os.Stdout)
				os.Exit(0)
			}
		}
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
