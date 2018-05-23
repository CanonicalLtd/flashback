// -*- Mode: Go; indent-tabs-mode: t -*-
// Curtin Core
// Copyright 2018 Canonical Ltd.  All rights reserved.

package curtin

import "fmt"

func diskHandler(blockMeta PartitionCommand, item StorageItem) error {
	fmt.Printf("Initialize the disk `%s`\n", item.Path)
	// Define the disk path using the different formats
	diskPath := DiskPath{
		Device:       deviceNameFromPath(item.Path),
		DevicePath:   item.Path,
		SysBlockPath: sysBlockFromPath(item.Path),
	}

	if item.Preserve {
		fmt.Println("  Not implemented: ignore the `preserve` flag")
	}

	if !item.Preserve && len(item.Wipe) > 0 {
		fmt.Println("  Not implemented: wipe the disk")
	}

	if !item.Preserve && len(item.PTable) > 0 {
		if item.PTable == pTableGPT {
			// Wipe both MBR and GPT that may be present on the disk.
			// n.b. wipe_volume wipes 1M at front and end of the disk.
			// This could destroy disk data in filesystems that lived
			// there.
			fmt.Printf("  Wipe the MBR and GPT on the disk `%s`\n", diskPath.Device)
			if err := wipeVolume(diskPath, wipeModeSuperblock, true); err != nil {
				fmt.Println(err)
				return err
			}
		}
		fmt.Println("  Not implemented: remove holders and label the disk")
	}

	return nil
}
