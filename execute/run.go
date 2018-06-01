// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package execute

import (
	"fmt"
	"strings"

	"github.com/CanonicalLtd/curtin-core/curtin"
	flags "github.com/jessevdk/go-flags"
)

// RunCommand defines the execution options for the application
type RunCommand struct {
	ConfigPath string `short:"c" long:"config" description:"read configuration from cfg"`
}

// Execute the run command
func (cmd RunCommand) Execute(args []string) error {
	// Read the config parameters
	config, err := curtin.ReadConfig(cmd.ConfigPath)
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
	partition(config)

	return nil
}

func partition(config curtin.Config) error {
	for _, v := range config.PartitioningCommands {
		args, err := flags.ParseArgs(&Execution, strings.Split(v, " ")[1:])
		if err != nil {
			return err
		}
		if len(args) == 0 {
			return fmt.Errorf("The `meta-mode` parameter must be provided")
		}
		// Set the mode from the first arg
		Execution.BlockMeta.Mode = args[0]

		// Run the partition command
		err = curtin.Partition(Execution.BlockMeta, config)
		if err != nil {
			return err
		}
	}
	return nil
}
