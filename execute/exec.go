// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package execute

import "github.com/CanonicalLtd/flashback/core"

// Command defines the execution options for the application
type Command struct {
	Run       RunCommand            `command:"run" alias:"r" description:"Execute the curt installer"`
	BlockMeta core.PartitionCommand `command:"block-meta" description:"Partition the storage"`
}

// Execution is the implementation of the execution options
var Execution Command
