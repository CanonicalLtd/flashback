// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/CanonicalLtd/flashback/core"
	"github.com/CanonicalLtd/flashback/execute"
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
	config, err := core.ReadConfig(execute.Execution.Run.ConfigPath)
	if err != nil {
		return err
	}

	// Inform of ignored commands in the config file
	if len(config.EarlyCommands) > 0 {
		fmt.Println("Ignoring `early_commands` - not supported")
	}
	if len(config.NetworkCommands) > 0 {
		fmt.Println("Ignoring `network_commands` - not supported")
	}

	// Execute the partition commands
	return execPartitionCommands(config)
}

func execPartitionCommands(config core.Config) error {
	for _, v := range config.PartitioningCommands {
		// The partition commands are in the format `curtin block-meta --target=/mnt/here simple`
		// Curtin uses the command to self-execute, but we just want the mode and parameters
		args, err := flags.ParseArgs(&execute.Execution, strings.Split(v, " ")[1:])
		if err != nil {
			return err
		}
		if len(args) == 0 {
			return fmt.Errorf("The `meta-mode` parameter must be provided")
		}
		if len(args) != 2 {
			return fmt.Errorf("The `meta-mode` must be in the format `block-meta simple|custom`")
		}

		// Set the mode from the second arg e.g. block-meta >>simple<<
		execute.Execution.BlockMeta.Mode = args[1]

		// Run the partition command
		err = core.Partition(execute.Execution.BlockMeta, config)
		if err != nil {
			return err
		}
	}
	return nil
}
