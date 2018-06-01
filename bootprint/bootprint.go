// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package bootprint

import (
	"fmt"

	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/config"
	"github.com/CanonicalLtd/flashback/core"
)

// CheckAndRun verifies that a restore partition has been created
// If not, it initiates the creation of the restore partition
func CheckAndRun() (string, error) {
	device, err := core.FindFS(config.Store.RestorePartitionLabel)
	if err == nil && len(device) > 0 {
		// We have completed the boot print previously
		return device, nil
	}

	// Looks as though the restore partition has not been created... let's take a boot print!
	return Run()
}

// Run executes the backup of the initial writable partition and system-boot data
func Run() (string, error) {
	// Find original "writable" and matching disk device
	writable, err := core.FindFS(config.Store.WritablePartitionLabel)
	if err != nil {
		audit.Printf("Cannot find the writable partition: `%s` : %v\n", config.Store.WritablePartitionLabel, err)
		return "", nil
	}
	fmt.Println("---", writable)

	// TODO: Set the clock to image creation time so we are not too far off

	// find free partition space
	last, err := core.FreePartitionSpace(writable)
	fmt.Println("---", last, err)

	// re-label old "writable" to "restore"
	a, err := core.RelabelDisk(writable, config.Store.RestorePartitionLabel)
	fmt.Println("---", a, err)

	// check which partition table we use

	// refresh partition table

	// # encrypt new partition

	// format the device

	// back up system-boot

	// # mark superblock of restore partition readonly
	return "", nil
}
