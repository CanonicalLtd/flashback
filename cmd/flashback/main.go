// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package main

import (
	"os"

	"github.com/CanonicalLtd/flashback/audit"

	"github.com/CanonicalLtd/flashback/bootprint"
	"github.com/CanonicalLtd/flashback/config"
	"github.com/CanonicalLtd/flashback/execute"
	"github.com/CanonicalLtd/flashback/reset"
	flags "github.com/jessevdk/go-flags"
)

func main() {
	args, err := flags.ParseArgs(&execute.Execution, os.Args)
	if err != nil {
		os.Exit(1)
	}

	err = Execute(args)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

// Execute processes the args and runs the image restore
func Execute(args []string) error {
	// Read the config parameters
	err := config.Read(execute.Execution.ConfigPath)
	if err != nil {
		return err
	}

	// Check if we need to create a boot print
	err = bootprint.CheckAndRun()
	if err != nil {
		return err
	}

	// Start a factory reset, if requested
	if execute.Execution.FactoryReset {
		err = reset.Run()
		if err != nil {
			audit.Println("Error in factory reset:", err)
		}
	}

	return err
}
