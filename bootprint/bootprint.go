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

	// // TODO: find free partition space
	// last, err := core.FreePartitionSpace(writable)
	// encryptPart := last + 1
	// encryptDevice := core.DevicePathFromNumber(writable, encryptPart)

	// fmt.Println("---last partition", last, err)
	// fmt.Println("---encrypt", encryptPart, encryptDevice)

	// re-label old "writable" to "restore"
	a, err := core.RelabelDisk(writable, config.Store.RestorePartitionLabel)
	fmt.Println("---relabel", a, err)

	// Create the `writable` partition with the free space
	newWritable, err := core.CreateNextPartition(writable)
	if err != nil {
		return "", err
	}

	// Refresh partition table, ignore error as the device may be busy
	_ = core.RefreshPartitionTable(writable)

	// # encrypt new partition

	// Format the new partition
	err = core.FormatDisk(newWritable, "ext4", config.Store.WritablePartitionLabel)
	if err != nil {
		return newWritable, err
	}

	fmt.Println("---format", err)

	// back up system-boot
	err = core.BackupSystemBoot(config.Store.BootPartitionLabel, config.Store.RestorePartitionLabel)
	if err != nil {
		return newWritable, err
	}

	// # mark superblock of restore partition readonly
	return newWritable, nil
}
