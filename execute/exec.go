// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package execute

import "github.com/CanonicalLtd/flashback/core"

// RunCommand defines the execution options for the application
type RunCommand struct {
	ConfigPath string `short:"c" long:"config" description:"read configuration from cfg"`
}

// Command defines the execution options for the application
type Command struct {
	Run       RunCommand
	BlockMeta core.PartitionCommand
}

// Execution is the implementation of the execution options
var Execution Command
