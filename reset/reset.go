// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package reset

import (
	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/core"
)

// Run starts the factory reset
func Run() error {
	audit.Println("Start a factory reset of the device")

	// Find the partition devices
	err := core.FindPartitions()
	if err != nil {
		return nil
	}

	// Create a RAM disk copy of the restore partition
	if err := createRestoreRAMDisk(); err != nil {
		return err
	}

	// Back up the requested data to the RAM disk copy of restore
	if err := backupUserData(); err != nil {
		audit.Println("Error backing up user data to copy of `restore` partition")
		return err
	}

	// Write the RAM disk to the restore partition
	if err := copyRAMDiskToRestore(); err != nil {
		audit.Println("Error copying user data to `restore` partition")
		return err
	}

	// // Restore writable to virgin state

	// // Restore system-boot to virgin state
	// audit.Println("Restore system-boot to its first-boot state")
	// if err = restoreSystemBoot(config.Store.RestorePartitionLabel, config.Store.BootPartitionLabel); err != nil {
	// 	return err
	// }

	// // Restore backed up data
	// if err := restoreUserData(writable); err != nil {
	// 	return err
	// }

	_ = core.Unmount(core.TargetPath)
	_ = core.Unmount(core.RestorePath)
	_ = core.Unmount(core.TempFSMount)

	// Initiate reboot
	return nil
}
