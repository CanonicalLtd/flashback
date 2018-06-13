// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package reset

import (
	"github.com/CanonicalLtd/flashback/core"
)

// restoreSystemBoot restores system-boot from the raw backup
func restoreSystemBoot() error {
	// Mount the restore path
	err := core.Mount(core.PartitionTable.Restore, core.RestorePath)
	if err != nil {
		return err
	}

	// Unmount the boot path
	_ = core.Unmount(core.PartitionTable.SystemBoot)

	// Write partition content back
	err = core.UnzipToDevice(core.BackupImageSystemBoot, core.PartitionTable.SystemBoot)

	// Unmount the restore partition
	_ = core.Unmount(core.RestorePath)

	return err
}
