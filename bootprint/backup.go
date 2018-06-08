// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package bootprint

import (
	"github.com/CanonicalLtd/flashback/core"
)

// backup makes a raw backup of a partition
func backupPartition(devicePath, imagePath string) error {
	// Mount the restore path
	err := core.Mount(core.PartitionTable.Restore, core.RestorePath)
	if err != nil {
		return err
	}

	// Unmount the boot path
	_ = core.Unmount(devicePath)

	// Back up the partition to img file so we keep the exact filesystem
	// without having to parse gadget.yaml or worrying about ABI compatibility
	// to ubuntu-image's dosfstools
	err = core.ReadAndGzipToFile(devicePath, imagePath)

	// Unmount the restore partition
	_ = core.Unmount(core.RestorePath)

	return err
}

// backupWritable makes a raw backup of writable partition
func backupWritable() error {
	return backupPartition(core.PartitionTable.Writable, core.BackupImageWritable)
}

// backupSystemBoot makes a raw backup of system-boot partition
func backupSystemBoot() error {
	return backupPartition(core.PartitionTable.SystemBoot, core.BackupImageSystemBoot)
}
