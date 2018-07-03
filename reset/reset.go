// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package reset

import (
	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/config"
	"github.com/CanonicalLtd/flashback/core"
)

// Run starts the factory reset
func Run() error {
	audit.Println("Start a factory reset of the device")

	// Find the partition devices
	err := core.FindPartitions()
	if err != nil {
		return err
	}

	// Create a RAM disk copy of the restore partition
	if err := core.CreateTmpfsDisk(core.TempFSMount, config.Store.Backup.Size); err != nil {
		return err
	}

	// Back up the requested data to the RAM disk copy of restore
	if err := backupUserData(); err != nil {
		audit.Println("Error backing up user data to copy of `restore` partition")
		return err
	}

	// Get the partition type of writable
	fsType, err := core.FSType(core.PartitionTable.Writable)
	if err != nil {
		return err
	}

	// Format the writable partition
	_ = core.Unmount(core.PartitionTable.Writable)
	if err := core.FormatDisk(core.PartitionTable.Writable, fsType, core.PartitionWritable); err != nil {
		audit.Println("Error formatting the `writable` partition")
		return err
	}

	// Restore writable from the backup file on the restore partition
	if err := restoreWritable(); err != nil {
		audit.Println("Error restoring the `writable` partition")
		return err
	}

	// Restore system-boot to virgin state by rewriting the partition from the backup
	audit.Println("Restore system-boot to its first-boot state")
	if err = restoreSystemBoot(); err != nil {
		return err
	}

	// Restore backed up data
	if err := restoreUserData(); err != nil {
		return err
	}

	audit.Println("Factory reset completed successfully")

	_ = core.Unmount(core.WritablePath)
	_ = core.Unmount(core.RestorePath)
	_ = core.Unmount(core.TempFSMount)

	// Initiate reboot
	return nil
}
