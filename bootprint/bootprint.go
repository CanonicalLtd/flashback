// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package bootprint

import (
	"os"

	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/core"
)

// CheckAndRun verifies that a restore partition has been created
// If not, it initiates the creation of the restore partition
func CheckAndRun(check bool) error {
	// Find the partition devices
	err := core.FindPartitions()
	if err != nil {
		return nil
	}

	// Mount the restore path
	if err = core.Mount(core.PartitionTable.Restore, core.RestorePath); err != nil {
		return err
	}

	if check {
		// Check that the backup files exist
		backupBoot := false
		backupWritable := false
		if _, err := os.Stat(core.BackupImageSystemBoot); err == nil {
			backupBoot = true
		}
		if _, err := os.Stat(core.BackupImageWritable); err == nil {
			backupWritable = true
		}
		if backupBoot && backupWritable {
			audit.Println("Recovery image is already created")
			_ = core.Unmount(core.RestorePath)
			return nil
		}
	}

	// Looks as though the backup has not been created... let's take a boot print!
	return Run()
}

// Run executes the backup of the initial writable partition and system-boot data
func Run() error {
	audit.Println("Create the recovery image")
	// TODO: Set the clock to image creation time so we are not too far off

	// Back up writable
	audit.Println("Backup the writable partition")
	if err := backupWritable(); err != nil {
		return err
	}

	// Back up system-boot
	audit.Println("Backup the system boot partition")
	if err := backupSystemBoot(); err != nil {
		return err
	}

	// # mark superblock of restore partition readonly
	return nil
}
