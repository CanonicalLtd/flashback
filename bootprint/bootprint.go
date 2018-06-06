// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package bootprint

import (
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
	audit.Printf("Find the writable partition: %s", config.Store.WritablePartitionLabel)
	writable, err := core.FindFS(config.Store.WritablePartitionLabel)
	if err != nil {
		audit.Printf("Cannot find the writable partition: `%s` : %v\n", config.Store.WritablePartitionLabel, err)
		return "", nil
	}
	audit.Println("Found partition at", writable)

	// TODO: Set the clock to image creation time so we are not too far off

	// // TODO: find free partition space
	// last, err := core.FreePartitionSpace(writable)
	// encryptPart := last + 1
	// encryptDevice := core.DevicePathFromNumber(writable, encryptPart)

	// fmt.Println("---last partition", last, err)
	// fmt.Println("---encrypt", encryptPart, encryptDevice)

	// Re-label old "writable" to "restore"
	audit.Println("Relabel the writable partition as", config.Store.RestorePartitionLabel)
	resp, err := core.RelabelDisk(writable, config.Store.RestorePartitionLabel)
	if err != nil {
		audit.Println(resp)
		audit.Printf("Cannot find the relabel partition: `%s` : %v\n", config.Store.RestorePartitionLabel, err)
		return "", nil
	}

	// Create the `writable` partition with the free space
	audit.Println("Create the writable partition with the free space")
	newWritable, err := core.CreateNextPartition(writable)
	if err != nil {
		return "", err
	}

	// Refresh partition table, ignore error as the device may be busy
	_ = core.RefreshPartitionTable(writable)

	// # encrypt new partition

	// Format the new partition
	audit.Println("Format the writable partition:", newWritable)
	if err = core.FormatDisk(newWritable, "ext4", config.Store.WritablePartitionLabel); err != nil {
		return newWritable, err
	}

	// Copy content from restore partition (renamed writable) to the new writable partition
	audit.Println("Copy the system data to the new partition")
	if err = core.CopySystemData(writable, newWritable); err != nil {
		return newWritable, err
	}

	// back up system-boot
	audit.Println("Backup the system boot partition")
	if err = backupSystemBoot(config.Store.BootPartitionLabel, config.Store.RestorePartitionLabel); err != nil {
		return newWritable, err
	}

	// # mark superblock of restore partition readonly
	return newWritable, nil
}
