// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package execute

import "github.com/CanonicalLtd/curtin-core/curtin"

// Command defines the execution options for the application
type Command struct {
	Run       RunCommand              `command:"run" alias:"r" description:"Execute the curt installer"`
	BlockMeta curtin.PartitionCommand `command:"block-meta" description:"Partition the storage"`
}

// Execution is the implementation of the execution options
var Execution Command
