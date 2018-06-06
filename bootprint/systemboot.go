// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package bootprint

import (
	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/core"
)

// backupSystemBoot makes a raw backup of system-boot
func backupSystemBoot(systemBoot, restore string) error {
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

	// Back up system-boot partition to img file so we keep the exact filesystem
	// without having to parse gadget.yaml or worrying about ABI compatibility
	// to ubuntu-image's dosfstools
	err = core.ReadAndGzipToFile(deviceBoot, core.BackupImagePath)

	// Unmount the restore partition
	_ = core.Unmount(core.RestorePath)

	return err
}
