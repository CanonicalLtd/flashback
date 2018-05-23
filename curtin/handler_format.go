// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package curtin

import (
	"errors"
	"fmt"
)

func formatHandler(blockMeta PartitionCommand, item StorageItem) error {
	fmt.Printf("Format the partition `%s`\n", item.Volume)

	// Check the format details
	if len(item.Volume) == 0 {
		return errors.New("volume must be set for partition to be formatted")
	}
	if item.Preserve {
		fmt.Printf("  Partition `%s` is set to be preserved - skip formatting\n", item.Volume)
		return nil
	}

	// Define the partition path using the different formats
	ptnPath := DiskPath{
		Device:       item.Volume,
		DevicePath:   devicePathFromDevice(item.Volume),
		SysBlockPath: sysBlockFromDevice(item.Volume),
	}

	err := mkfs(ptnPath.DevicePath, item.FsType, item.Label)
	return err
}
