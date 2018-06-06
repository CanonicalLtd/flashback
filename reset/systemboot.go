// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package reset

import (
	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/core"
)

// restoreSystemBoot restores system-boot from the raw backup
func restoreSystemBoot(restore, systemBoot string) error {
	// Get the boot and restore partitions
	deviceBoot, err := core.FindFS(systemBoot)
	if err != nil {
		audit.Printf("Cannot find the `%s` partition: %v", systemBoot, err)
		return err
	}
	deviceRestore, err := core.FindFS(restore)
	if err != nil {
		audit.Printf("Cannot find the `%s` partition: %v", restore, err)
		return err
	}

	// Mount the restore path
	err = core.Mount(deviceRestore, core.RestorePath)
	if err != nil {
		return err
	}

	// Unmount the boot path
	_ = core.Unmount(deviceBoot)

	// Write partition content back
	err = core.UnzipToDevice(core.BackupImagePath, deviceBoot)

	// Unmount the restore partition
	_ = core.Unmount(core.RestorePath)

	return err
}
